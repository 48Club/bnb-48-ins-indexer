package servers

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/jwrookie/fans/config"
	"github.com/jwrookie/fans/dao"
	"github.com/jwrookie/fans/pkg/database"
	"github.com/jwrookie/fans/pkg/global"
	"github.com/jwrookie/fans/pkg/log"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"github.com/jwrookie/fans/pkg/utils"
	"gorm.io/gorm"
	"math/big"
	"strings"
	"time"
)

type BscScanService struct {
	account        dao.IAccount
	accountRecords dao.IAccountRecords
	accountWallet  dao.IAccountWallet
	inscriptionDao dao.IInscription
	inscriptions   map[string]inscription // tick-hash : inscription
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
		inscriptions:   make(map[string]inscription),
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

		s.inscriptions[ele.TickHash] = insc
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
	input := common.Bytes2Hex(tx.Data())
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
	}

	return nil
}

func (s *BscScanService) deploy(db *gorm.DB, block *types.Block, tx *types.Transaction, insc *bnb48types.BNB48Inscription, index int) error {
	return nil
}

func (s *BscScanService) recap() error {
	return nil
}

func (s *BscScanService) mint(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *bnb48types.BNB48Inscription, index int) error {
	lim, err := utils.StringToBigint(inscription.Lim)
	if err != nil {
		log.Sugar.Error(tx.Hash().Hex(), err)
		return nil
	}

	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		return nil
	}
	// minting ended
	if insc.Minted.Cmp(insc.Max) >= 0 {
		return nil
	}
	// lim
	if lim.Cmp(insc.Lim) > 0 || lim.Cmp(big.NewInt(0)) < 0 {
		return nil
	}
	// miners
	_, ok = insc.Miners[block.Coinbase().Hex()]
	if len(insc.Miners) > 0 && !ok {
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	// account
	account, err := s.account.SelectByAddress(db, from)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = &dao.AccountModel{Address: from}
			if account.Id, err = utils.GenSnowflakeID(); err != nil {
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
	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, account.Id, inscription.TickHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			accountWallet = &dao.AccountWalletModel{AccountId: account.Id, Tick: insc.Tick, TickHash: insc.TickHash}
			if accountWallet.Id, err = utils.GenSnowflakeID(); err != nil {
				return err
			}
			if err = s.accountWallet.Create(db, accountWallet); err != nil {
				return err
			}
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

	// add record
	record := &dao.AccountRecordsModel{
		Block:   block.NumberU64(),
		BlockAt: block.Time() * 1000,
		TxHash:  tx.Hash().Hex(),
		TxIndex: uint64(index),
		From:    from,
		To:      strings.ToLower(tx.To().Hex()),
		Input:   strings.ToLower("0x" + common.Bytes2Hex(tx.Data())),
		Type:    1, // mint
	}
	if record.Id, err = utils.GenSnowflakeID(); err != nil {
		return err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return err
	}

	return nil
}

func (s *BscScanService) transfer(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *bnb48types.BNB48Inscription, index int) error {
	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		return nil
	}

	if !utils.IsValidERCAddress(inscription.To) {
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

func (s *BscScanService) transferForFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *bnb48types.BNB48Inscription, index int) (*big.Int, error) {
	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	zero := new(big.Int)

	// sub balance of tx from
	fromAccount, err := s.account.SelectByAddress(db, from)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return zero, nil
		} else {
			return nil, err
		}
	}

	amt, err := utils.StringToBigint(inscription.Amt)
	if err != nil {
		log.Sugar.Error(tx.Hash().Hex(), err)
		return zero, nil
	}

	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, fromAccount.Id, inscription.TickHash)
	if err != nil {
		return nil, err
	}
	currentBalance, err := utils.StringToBigint(accountWallet.Balance)
	if currentBalance.Cmp(amt) < 0 {
		return zero, nil
	}

	updates := map[string]interface{}{
		"balance": new(big.Int).Sub(currentBalance, amt).String(),
	}
	if err = s.accountWallet.UpdateBalance(db, accountWallet.Id, updates); err != nil {
		return nil, err
	}

	// add record for tx from
	record := &dao.AccountRecordsModel{
		Block:   block.NumberU64(),
		TxHash:  tx.Hash().Hex(),
		TxIndex: uint64(index),
		From:    from,
		To:      strings.ToLower(tx.To().Hex()),
		Input:   strings.ToLower("0x" + common.Bytes2Hex(tx.Data())),
		Type:    2, // transfer
	}
	if record.Id, err = utils.GenSnowflakeID(); err != nil {
		return nil, err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return nil, err
	}

	return amt, nil
}

func (s *BscScanService) transferForTo(db *gorm.DB, amt *big.Int, insc inscription, to string) error {
	// account
	account, err := s.account.SelectByAddress(db, to)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = &dao.AccountModel{Address: to}
			if account.Id, err = utils.GenSnowflakeID(); err != nil {
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
	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, account.Id, insc.TickHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			accountWallet = &dao.AccountWalletModel{AccountId: account.Id, Tick: insc.Tick, TickHash: insc.TickHash}
			if accountWallet.Id, err = utils.GenSnowflakeID(); err != nil {
				return err
			}
			if err = s.accountWallet.Create(db, accountWallet); err != nil {
				return err
			}
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
