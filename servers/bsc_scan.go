package servers

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jwrookie/fans/config"
	"github.com/jwrookie/fans/dao"
	"github.com/jwrookie/fans/pkg/database"
	"github.com/jwrookie/fans/pkg/global"
	"github.com/jwrookie/fans/pkg/log"
	"github.com/jwrookie/fans/pkg/utils"
	"gorm.io/gorm"
	"strings"
	"time"
)

type BscScanService struct {
	account        dao.IAccount
	accountRecords dao.IAccountRecords
	accountHash    dao.IAccountHash
	conf           config.Config
}

func NewBscScanService() *BscScanService {
	return &BscScanService{
		account:        &dao.AccountHandler{},
		accountRecords: &dao.AccountRecordsHandler{},
		accountHash:    &dao.AccountHashHandler{},
		conf:           config.GetConfig(),
	}
}

func (s *BscScanService) Scan() error {
	block := s.conf.BscIndex.ScanBlock

	for {
		targetBlock, err := global.BscClient.BlockByNumber(context.TODO(), nil)
		if err != nil {
			return err
		}

		if targetBlock.NumberU64() < block {
			time.Sleep(time.Second)
		}

		if err = s.work(targetBlock); err != nil {
			return err
		}

		log.Sugar.Infof("currentScanBlock: %d\n", block)
		block++

		if err = config.SaveETHConfig(block); err != nil {
			return err
		}
	}
}

func (s *BscScanService) work(block *types.Block) error {
	db := database.Mysql().Begin()
	defer db.Rollback()
	for _, tx := range block.Transactions() {
		if err := s._work1(db, block.NumberU64(), tx, block.Coinbase().Hex()); err != nil {
			return err
		}
	}

	return db.Commit().Error
}

func (s *BscScanService) _work1(db *gorm.DB, blockNumber uint64, tx *types.Transaction, coinbase string) error {
	if s.conf.App.MintStartBlock <= blockNumber && blockNumber <= s.conf.App.MintEndBlock {
		if !strings.EqualFold(coinbase, global.BNB48) {
			return nil
		}

		return s.mint(db, tx, blockNumber)
	}

	if s.conf.App.MintEndBlock <= blockNumber && blockNumber <= s.conf.App.BridgeEvmBlock {
		return s.transfer(db, tx, blockNumber)
	}

	return nil
}

func (s *BscScanService) mint(db *gorm.DB, tx *types.Transaction, blockNumber uint64) error {
	input := "0x" + common.Bytes2Hex(tx.Data())
	if !strings.EqualFold(input, global.MintData) {
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	model, err := s.account.SelectByAddress(db, from)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model = &dao.AccountModel{}
			if model.Id, err = utils.GenSnowflakeID(); err != nil {
				return err
			}

			if err = s.account.Create(db, model); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	updates := map[string]interface{}{
		"balance": model.Balance + 1,
	}

	if err = s.account.UpdateBalance(db, model.Id, updates); err != nil {
		return err
	}

	record := &dao.AccountRecordsModel{
		Block:  blockNumber,
		TxHash: tx.Hash().Hex(),
		From:   from,
		To:     strings.ToLower(tx.To().Hex()),
		Input:  strings.ToLower(input),
		Type:   1, // mint
	}
	if record.Id, err = utils.GenSnowflakeID(); err != nil {
		return err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return err
	}

	hash := &dao.AccountHashModel{
		AccountId: model.Id,
		MintHash:  tx.Hash().Hex(),
		State:     1, // valid
	}
	if err = s.accountHash.Create(db, hash); err != nil {
		return err
	}

	return nil
}

func (s *BscScanService) transfer(db *gorm.DB, tx *types.Transaction, blockNumber uint64) error {
	input := "0x" + common.Bytes2Hex(tx.Data())
	if len(input) != 66 {
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	if strings.EqualFold(from, tx.To().Hex()) {
		return nil
	}

	// s1: select from account
	fromAccount, err := s.account.SelectByAddress(db, from)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	if fromAccount.Balance < 1 {
		return nil
	}

	// s2: make sure the hash has not been transferred
	fromHash, err := s.accountHash.Select(db, &dao.AccountHashModel{
		AccountId: fromAccount.Id,
		MintHash:  strings.ToLower(input),
		State:     1,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	// s3: update from balance
	formUpdates := map[string]interface{}{
		"balance": fromAccount.Balance - 1,
	}

	if err = s.account.UpdateBalance(db, fromAccount.Id, formUpdates); err != nil {
		return err
	}

	// s4: update to balance
	toModel, err := s.account.SelectByAddress(db, strings.ToLower(tx.To().Hex()))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			toModel = &dao.AccountModel{}
			if toModel.Id, err = utils.GenSnowflakeID(); err != nil {
				return err
			}

			if err = s.account.Create(db, toModel); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	toUpdates := map[string]interface{}{
		"balance": toModel.Balance + 1,
	}

	if err = s.account.UpdateBalance(db, toModel.Id, toUpdates); err != nil {
		return err
	}

	// s5: add to hash
	toHash := &dao.AccountHashModel{
		AccountId: toModel.Id,
		MintHash:  fromHash.MintHash,
		State:     1, // valid
	}
	if err = s.accountHash.Create(db, toHash); err != nil {
		return err
	}

	// s6: update from hash, make it invalid
	fromHashUpdates := map[string]interface{}{
		"state":     2, // invalid
		"delete_at": time.Now().UnixMilli(),
	}
	if err = s.accountHash.Update(db, fromHash.Id, fromHashUpdates); err != nil {
		return err
	}

	// s7: add from records
	record := &dao.AccountRecordsModel{
		Block:  blockNumber,
		TxHash: tx.Hash().Hex(),
		From:   from,
		To:     strings.ToLower(tx.To().Hex()),
		Input:  strings.ToLower(input),
		Type:   2, // transfer
	}
	if record.Id, err = utils.GenSnowflakeID(); err != nil {
		return err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return err
	}

	return nil
}
