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

	types2 "bnb-48-ins-indexer/pkg/types"

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
	pendingTxs     *types2.GlobalVariable
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

func NewBscScanService(pendingTxs *types2.GlobalVariable) *BscScanService {
	return &BscScanService{
		account:        &dao.AccountHandler{},
		accountRecords: &dao.AccountRecordsHandler{},
		accountWallet:  &dao.AccountWalletHandler{},
		inscriptionDao: &dao.InscriptionHandler{},
		inscriptions:   make(map[string]*inscription),
		conf:           config.GetConfig(),
		pendingTxs:     pendingTxs,
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

func (s *BscScanService) checkPendingTxs(beginBH *types.Header) {

	{
		// 删除已经确认的交易
		_tmpTxsInBlock := s.pendingTxs.TxsInBlock
		if _tmpTxsInBlock.Cardinality() > 0 {
			for _, bn := range s.pendingTxs.TxsInBlock.ToSlice() {
				if beginBH.Number.Uint64()-bn >= 15 {
					_tmpTxsInBlock.Remove(bn)
				}
			}
			s.pendingTxs.TxsInBlock = _tmpTxsInBlock
		}

		_tmpTxsByAddr := s.pendingTxs.TxsByAddr
		for addr, records := range _tmpTxsByAddr {
			for tk_hash, v := range records {
				_tmpRecords := v
				for _, v := range _tmpRecords {
					if s.pendingTxs.TxsInBlock.Contains(v.Block) {
						continue
					}
					if beginBH.Number.Uint64()-v.Block >= 15 {
						// delete record in s.pendingTxs
						delete(s.pendingTxs.Txs, v.TxHash)
						s.pendingTxs.TxsHash.Remove(v.TxHash)
						delete(_tmpTxsByAddr[addr][tk_hash], v.TxHash)
					}

				}
			}

		}
		s.pendingTxs.TxsByAddr = _tmpTxsByAddr
	}
	{
		// 添加新的交易
		targetBlockHeader, err := global.BscClient.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return
		}
		if s.pendingTxs.BlockAt.Cmp(common.Big0) == 0 {
			// 更新区块高度
			s.pendingTxs.BlockAt = beginBH.Number
		}
		for {
			targetBlock, err := global.BscClient.BlockByNumber(context.Background(), s.pendingTxs.BlockAt)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			if err = s.work(targetBlock, true); err != nil {
				return
			}
			if s.pendingTxs.BlockAt.Cmp(targetBlockHeader.Number) == 0 {
				break
			}

			s.pendingTxs.BlockAt = new(big.Int).Add(s.pendingTxs.BlockAt, common.Big1)
		}
	}

}
func (s *BscScanService) Scan() error {
	var err error
	if err = s.init(); err != nil {
		return err
	}

	block := s.conf.BscIndex.ScanBlock

	for {

		targetBN := new(big.Int).SetUint64(block)
		targetBlockHeader, err := global.BscClient.HeaderByNumber(context.Background(), targetBN)
		if err == nil {
			go s.checkPendingTxs(targetBlockHeader)
		}
		if err != nil || time.Now().Unix()-int64(targetBlockHeader.Time) < 45 {
			time.Sleep(time.Second)
			continue
		}

		targetBlock, err := global.BscClient.BlockByNumber(context.Background(), new(big.Int).SetUint64(block))

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

func (s *BscScanService) work(block *types.Block, isPending ...bool) error {
	db := database.Mysql().Begin()
	defer db.Rollback()

	for index, tx := range block.Transactions() {
		if err := s._work(db, block, tx, index, isPending...); err != nil {
			return err
		}
	}
	if len(isPending) > 0 {
		return db.Rollback().Error
	}

	return db.Commit().Error
}

func (s *BscScanService) _work(db *gorm.DB, block *types.Block, tx *types.Transaction, index int, isPending ...bool) error {
	data, err := utils.InputToBNB48Inscription(hexutil.Encode(tx.Data()))
	if err != nil {
		log.Sugar.Error(err)
		return nil
	}
	if data == nil {
		return nil
	}
	switch data.Op {
	case "deploy":
	case "recap":
	case "mint":
		if err = s.mint(db, block, tx, data, index, isPending...); err != nil {
			return err
		}
	case "transfer":
		if err = s.transfer(db, block, tx, data, index, isPending...); err != nil {
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

func (s *BscScanService) mint(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int, isPending ...bool) error {
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
	// add record

	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		BlockAt:  block.Time(),
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		TickHash: insc.TickHash,
		From:     from,
		To:       strings.ToLower(tx.To().Hex()),
		Input:    hexutil.Encode(tx.Data()),
		Type:     1, // mint
	}

	if len(isPending) > 0 && isPending[0] {
		s.updateRam(record, block)
	}

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
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, 1); err != nil {
			return err
		}
	} else if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, -1); err != nil {
			return err
		}
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

func (s *BscScanService) transfer(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int, isPending ...bool) error {
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

	amt, err := s.transferForFrom(db, block, tx, inscription, index, isPending...)
	if err != nil {
		return err
	}
	if amt.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if err = s.transferForTo(db, amt, insc, inscription.To, isPending...); err != nil {
		return err
	}

	return nil
}

func (s *BscScanService) transferForFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int, isPending ...bool) (*big.Int, error) {
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

	// add record for tx from
	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		TickHash: inscription.TickHash,
		BlockAt:  block.Time(),
		From:     from,
		To:       strings.ToLower(tx.To().Hex()),
		Input:    hexutil.Encode(tx.Data()),
		Type:     2, // transfer
	}

	if len(isPending) > 0 && isPending[0] {
		s.updateRam(record, block)
	}

	balance := new(big.Int).Sub(currentBalance, amt)
	updates := map[string]interface{}{
		"balance": balance.String(),
	}
	if err = s.accountWallet.UpdateBalance(db, accountWallet.Id, updates); err != nil {
		return nil, err
	}
	if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, -1); err != nil {
			return nil, err
		}
	}

	if record.Id, err = dao.GenSnowflakeID(); err != nil {
		return nil, err
	}

	if err = s.accountRecords.Create(db, record); err != nil {
		return nil, err
	}

	return amt, nil
}

func (s *BscScanService) transferForTo(db *gorm.DB, amt *big.Int, insc *inscription, to string, isPending ...bool) error {
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
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, 1); err != nil {
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

/*
	type GlobalVariable struct {
		Txs           RecordsModelByTxHash
		TxsHash       mapset.Set[string]
		TxsByTickHash map[string]RecordsModelByTxHash
		TxsInBlock    mapset.Set[uint64]
		TxsByAddr     map[string]map[string][]dao.AccountRecordsModel
		BlockAt       *big.Int
	}
*/
func (s *BscScanService) updateRam(record *dao.AccountRecordsModel, block *types.Block) {
	if s.pendingTxs.TxsInBlock.Contains(block.NumberU64()) {
		return
	}
	record.IsPending = true
	record.InputDecode, _ = utils.InputToBNB48Inscription(record.Input)

	s.pendingTxs.TxsInBlock.Add(block.NumberU64())
	s.pendingTxs.Txs[record.TxHash] = record
	s.pendingTxs.TxsHash.Add(record.TxHash)

	if _, ok := s.pendingTxs.TxsByAddr[record.From]; !ok {
		s.pendingTxs.TxsByAddr[record.From] = make(map[string]types2.RecordsModelByTxHash)
	}
	s.pendingTxs.TxsByAddr[record.From][record.TickHash][record.TxHash] = record
	if _, ok := s.pendingTxs.TxsByAddr[record.To]; !ok {
		s.pendingTxs.TxsByAddr[record.To] = make(map[string]types2.RecordsModelByTxHash)
	}
	s.pendingTxs.TxsByAddr[record.To][record.TickHash][record.TxHash] = record

	if _, ok := s.pendingTxs.TxsByTickHash[record.TickHash]; !ok {
		s.pendingTxs.TxsByTickHash[record.TickHash] = make(map[string]*dao.AccountRecordsModel)
	}
	s.pendingTxs.TxsByTickHash[record.TickHash][record.TxHash] = record

}
