package service

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/global"
	"bnb-48-ins-indexer/pkg/helper"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

type BscScanService struct {
	account        dao.IAccount
	accountRecords dao.IAccountRecords
	accountWallet  dao.IAccountWallet
	inscriptionDao dao.IInscription
	inscriptions   map[string]*inscription // tick-hash : inscription
	conf           config.Config
}

type inscription struct {
	Id       uint64
	Miners   map[string]struct{}
	Max      *big.Int
	Lim      *big.Int
	Minted   *big.Int
	Tick     string
	TickHash string
}

func NewBscScanService() *BscScanService {
	return &BscScanService{
		account:        &dao.AccountHandler{},
		accountRecords: &dao.AccountRecordsHandler{},
		accountWallet:  &dao.AccountWalletHandler{},
		inscriptionDao: &dao.InscriptionHandler{},
		inscriptions:   make(map[string]*inscription),
		conf:           config.GetConfig(),
	}
}

func (s *BscScanService) init() error {
	inscs, err := s.inscriptionDao.Find(database.Mysql())
	if err != nil {
		return err
	}

	for _, ele := range inscs {
		var (
			insc   inscription
			miners = make(map[string]struct{})
		)
		insc.Id = ele.Id
		insc.Tick = ele.Tick
		insc.TickHash = ele.TickHash
		if insc.Max, err = utils.StringToBigint(ele.Max); err != nil {
			return err
		}
		if insc.Minted, err = utils.StringToBigint(ele.Minted); err != nil {
			return err
		}
		if insc.Lim, err = utils.StringToBigint(ele.Lim); err != nil {
			return err
		}

		for _, ele := range strings.Split(ele.Miners, ",") {
			miners[ele] = struct{}{}
		}

		s.inscriptions[ele.TickHash] = &insc
	}

	return nil
}

func (s *BscScanService) Scan() error {
	var err error
	if err = s.init(); err != nil {
		return err
	}

	block := s.conf.BscIndex.ScanBlock

	for {
		targetBlock, err := global.BscClient.BlockByNumber(context.TODO(), big.NewInt(int64(block)))
		if err != nil {
			if errors.Is(err, ethereum.NotFound) {
				time.Sleep(time.Second)
				continue
			} else {
				return err
			}
		}

		if err = s.work(targetBlock); err != nil {
			return err
		}

		log.Sugar.Infof("currentScanBlock: %d\n", block)
		block++

		if err = config.SaveBSCConfig(block); err != nil {
			return err
		}
	}
}

func (s *BscScanService) work(block *types.Block) error {
	db := database.Mysql().Begin()
	defer db.Rollback()

	for index, tx := range block.Transactions() {
		if err := s._work(db, block, tx, index); err != nil {
			return err
		}
	}

	return db.Commit().Error
}

func (s *BscScanService) _work(db *gorm.DB, block *types.Block, tx *types.Transaction, index int) error {
	input := "0x" + common.Bytes2Hex(tx.Data())
	if !strings.HasPrefix(input, global.BNB48Prefix) {
		return nil
	}

	data, err := utils.InputToBNB48Inscription(input)
	if err != nil {
		log.Sugar.Error(err)
		return nil
	}

	switch data.Op {
	case "deploy":
	case "recap":
	case "mint":
		if err = s.mint(db, block, tx, data, index); err != nil {
			return err
		}
	case "transfer":
		if err = s.transfer(db, block, tx, data, index); err != nil {
			return err
		}
	case "burn":
	case "approve":
	case "transferFrom":
	default:
		log.Sugar.Debugf("tx: %s, error: can not support %s op", tx.Hash().Hex(), data.Op)
	}

	return nil
}

func (s *BscScanService) deploy() error {
	return nil
}

func (s *BscScanService) recap() error {
	return nil
}

func (s *BscScanService) mint(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int) error {
	amt, err := utils.StringToBigint(inscription.Amt)
	if err != nil {
		log.Sugar.Debugf("tx: %s, error: %s, amt: %s", tx.Hash().Hex(), "amt invalid", inscription.Amt)
		return nil
	}

	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "not deploy")
		return nil
	}
	// minting ended
	if insc.Minted.Cmp(insc.Max) >= 0 {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "minting ended")
		return nil
	}
	// minting overflow
	if new(big.Int).Add(amt, insc.Minted).Cmp(insc.Max) > 0 {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "minting overflow")
		return nil
	}
	// amt
	if amt.Cmp(insc.Lim) > 0 || amt.Cmp(big.NewInt(0)) <= 0 {
		log.Sugar.Debugf("tx: %s, error: %s, want: 0 < amt < %d, get: %d", tx.Hash().Hex(), "amt invalid", insc.Lim, amt)
		return nil
	}
	// miners
	_, ok = insc.Miners[strings.ToLower(block.Coinbase().Hex())]
	if len(insc.Miners) > 0 && !ok {
		log.Sugar.Debugf("tx: %s, error: %s, want: %s, get: %s", tx.Hash().Hex(), "miners", insc.Miners, block.Coinbase().Hex())
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	// account
	account, err := s.account.SelectByAddress(db, from)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = &dao.AccountModel{Address: from}
			if account.Id, err = dao.GenSnowflakeID(); err != nil {
				return err
			}
			if err = s.account.Create(db, account); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// update accountWallet
	newWallet := false
	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, account.Id, inscription.TickHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			accountWallet = &dao.AccountWalletModel{AccountId: account.Id, Tick: insc.Tick, TickHash: insc.TickHash, Address: account.Address}
			if accountWallet.Id, err = dao.GenSnowflakeID(); err != nil {
				return err
			}
			if err = s.accountWallet.Create(db, accountWallet); err != nil {
				return err
			}
			newWallet = true
		} else {
			return err
		}
	}

	balance, err := utils.StringToBigint(accountWallet.Balance)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"balance": new(big.Int).Add(balance, big.NewInt(1)).String(),
	}
	if err = s.accountWallet.UpdateBalance(db, accountWallet.Id, updates); err != nil {
		return err
	}
	if newWallet {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.Tick, 1); err != nil {
			return err
		}
	} else if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.Tick, -1); err != nil {
			return err
		}
	}

	// add record
	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		BlockAt:  block.Time() * 1000,
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		TickHash: insc.TickHash,
		From:     from,
		To:       strings.ToLower(tx.To().Hex()),
		Input:    hexutil.Encode(tx.Data()),
		Type:     1, // mint
	}
	if record.Id, err = dao.GenSnowflakeID(); err != nil {
		return err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return err
	}

	// update inscription
	insc.Minted = new(big.Int).Add(insc.Minted, amt)
	inscUpdate := map[string]interface{}{
		"minted": insc.Minted.String(),
	}
	if insc.Minted.Cmp(insc.Max) == 0 {
		inscUpdate["status"] = 2 // mint end
	}
	if err = s.inscriptionDao.Update(db, insc.Id, inscUpdate); err != nil {
		return err
	}

	return nil
}

func (s *BscScanService) transfer(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int) error {
	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s, tick-hash: %s", tx.Hash().Hex(), "not deploy", inscription.TickHash)
		return nil
	}

	if !utils.IsValidERCAddress(inscription.To) {
		log.Sugar.Debugf("tx: %s, error: %s, to: %s", tx.Hash().Hex(), "invalid to", inscription.To)
		return nil
	}

	amt, err := s.transferForFrom(db, block, tx, inscription, index)
	if err != nil {
		return err
	}
	if amt.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if err = s.transferForTo(db, amt, insc, inscription.To); err != nil {
		return err
	}

	return nil
}

func (s *BscScanService) transferForFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int) (*big.Int, error) {
	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	zero := new(big.Int)

	amt, err := utils.StringToBigint(inscription.Amt)
	if err != nil {
		log.Sugar.Debugf("tx: %s, error: %s, amt: %s", tx.Hash().Hex(), "invalid amt", inscription.Amt)
		return zero, nil
	}

	if amt.Cmp(big.NewInt(0)) == 0 {
		log.Sugar.Debugf("tx: %s, error: %s, amt: %s", tx.Hash().Hex(), "invaild amt", "0")
		return zero, nil
	}

	// sub balance of tx from
	fromAccount, err := s.account.SelectByAddress(db, from)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "from account not found")
			return zero, nil
		} else {
			return nil, err
		}
	}

	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, fromAccount.Id, inscription.TickHash)
	if err != nil {
		return nil, err
	}
	currentBalance, err := utils.StringToBigint(accountWallet.Balance)
	if err != nil {
		return nil, err
	}
	if currentBalance.Cmp(amt) < 0 {
		log.Sugar.Debugf("tx: %s, error: %s, current balance: %s, need balance:%s", tx.Hash().Hex(), "insufficient balance", accountWallet.Balance, inscription.Amt)
		return zero, nil
	}

	balance := new(big.Int).Sub(currentBalance, amt)
	updates := map[string]interface{}{
		"balance": balance.String(),
	}
	if err = s.accountWallet.UpdateBalance(db, accountWallet.Id, updates); err != nil {
		return nil, err
	}
	if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.Tick, -1); err != nil {
			return nil, err
		}
	}

	// add record for tx from
	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		TickHash: inscription.TickHash,
		From:     from,
		To:       strings.ToLower(tx.To().Hex()),
		Input:    hexutil.Encode(tx.Data()),
		Type:     2, // transfer
	}
	if record.Id, err = dao.GenSnowflakeID(); err != nil {
		return nil, err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return nil, err
	}

	return amt, nil
}

func (s *BscScanService) transferForTo(db *gorm.DB, amt *big.Int, insc *inscription, to string) error {
	// account
	account, err := s.account.SelectByAddress(db, to)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = &dao.AccountModel{Address: to}
			if account.Id, err = dao.GenSnowflakeID(); err != nil {
				return err
			}
			if err = s.account.Create(db, account); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// update accountWallet
	newWallet := false
	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, account.Id, insc.TickHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			accountWallet = &dao.AccountWalletModel{AccountId: account.Id, Tick: insc.Tick, TickHash: insc.TickHash, Address: account.Address}
			if accountWallet.Id, err = dao.GenSnowflakeID(); err != nil {
				return err
			}
			if err = s.accountWallet.Create(db, accountWallet); err != nil {
				return err
			}
			newWallet = true
		} else {
			return err
		}
	}

	balance, err := utils.StringToBigint(accountWallet.Balance)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"balance": new(big.Int).Add(balance, amt).String(),
	}
	if err = s.accountWallet.UpdateBalance(db, accountWallet.Id, updates); err != nil {
		return err
	}
	if newWallet {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.Tick, 1); err != nil {
			return err
		}
	}
	return nil
}

func (s *BscScanService) burn() error {
	return nil
}

func (s *BscScanService) approve() error {
	return nil
}

func (s *BscScanService) transferFrom() error {
	return nil
}
