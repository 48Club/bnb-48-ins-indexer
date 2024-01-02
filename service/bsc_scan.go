package service

import (
	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/global"
	"bnb-48-ins-indexer/pkg/helper"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/utils"
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	types2 "bnb-48-ins-indexer/pkg/types"

	mapset "github.com/deckarep/golang-set/v2"
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
	Miners   mapset.Set[string]
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
			miners = mapset.NewSet[string]()
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

		if len(ele.Miners) >= 42 {
			for _, ele := range strings.Split(ele.Miners, ",") {
				miners.Add(ele)
			}
		}

		insc.Miners = miners

		s.inscriptions[ele.TickHash] = &insc
	}

	return nil
}

func (s *BscScanService) checkPendingTxs(beginBH *types.Header) {

	{
		s.pendingTxs.Lock()
		// 删除已经确认的交易
		for _, txHash := range s.pendingTxs.TxsHash.ToSlice() {
			if tx, ok := s.pendingTxs.Txs[txHash]; ok && beginBH.Number.Uint64() >= tx.Block {
				s.pendingTxs.TxsHash.Remove(txHash)
				delete(s.pendingTxs.Txs, txHash)
				if _, ok := s.pendingTxs.TxsByAddr[tx.From]; ok {
					if _, ok := s.pendingTxs.TxsByAddr[tx.From][tx.TickHash]; ok {
						delete(s.pendingTxs.TxsByAddr[tx.From][tx.TickHash], txHash)
					}
				}
				if _, ok := s.pendingTxs.TxsByAddr[tx.To]; ok {
					if _, ok := s.pendingTxs.TxsByAddr[tx.To][tx.TickHash]; ok {
						delete(s.pendingTxs.TxsByAddr[tx.To][tx.TickHash], txHash)
					}
				}
				if _, ok := s.pendingTxs.TxsByTickHash[tx.TickHash]; ok {
					delete(s.pendingTxs.TxsByTickHash[tx.TickHash], txHash)
				}
				s.pendingTxs.TxsInBlock.Remove(tx.Block)
			}
		}
		s.pendingTxs.Unlock()
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
		if err == nil && time.Now().Unix()-int64(targetBlockHeader.Time) < 45 {
			go s.checkPendingTxs(targetBlockHeader)
		}
		if err != nil || time.Now().Unix()-int64(targetBlockHeader.Time) < 45 {
			time.Sleep(time.Second)
			continue
		}

		targetBlock, err := global.BscClient.BlockByNumber(context.Background(), targetBN)

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
		// 当索引出现错误时, 需要回退区块重新同步需要添加 sp 事务
		// db.SavePoint("sp1")

		if err := s._work(db, block, tx, index, isPending...); err != nil {
			// if strings.Contains(strings.ToLower(err.Error()), strings.ToLower("1062 (23000): Duplicate entry")) {
			// 	db.RollbackTo("sp1")
			// 	continue
			// }
			return err
		}
	}
	if len(isPending) > 0 {
		return db.Rollback().Error
	}

	return db.Commit().Error
}

func (s *BscScanService) _work(db *gorm.DB, block *types.Block, tx *types.Transaction, index int, isPending ...bool) error {
	datas, err := utils.InputToBNB48Inscription(hexutil.Encode(tx.Data()), block.NumberU64())
	if err != nil {
		log.Sugar.Error(err)
		return nil
	}
	for opIndex, data := range datas {
		if data == nil {
			return nil
		}

		if data.P != "bnb-48" {
			return nil
		}

		switch data.Op {
		case "deploy":
			if err = s.deploy(db, block, tx, data, index, isPending...); err != nil {
				return err
			}
		case "recap":
		case "mint":
			if err = s.mint(db, block, tx, data, index, isPending...); err != nil {
				return err
			}
		case "transfer":
			if err = s.transfer(db, block, tx, data, index, opIndex, isPending...); err != nil {
				return err
			}
		case "burn":
			if err = s.burn(db, block, tx, data, index, opIndex, isPending...); err != nil {
				return err
			}
		case "approve":
		case "transferFrom":
		default:
			log.Sugar.Debugf("tx: %s, error: can not support %s op", tx.Hash().Hex(), data.Op)
		}
	}

	return nil
}

func (s *BscScanService) deploy(db *gorm.DB, block *types.Block, tx *types.Transaction, insc *helper.BNB48Inscription, index int, isPending ...bool) error {
	if len(isPending) > 0 {
		return nil
	}
	if insc.Tick == "" {
		log.Sugar.Debugf("tx: %s, error: %s, decimals: %s", tx.Hash().Hex(), "decimals invalid", insc.Decimals)
		return nil
	}

	decimals, err := utils.StringToBigint(insc.Decimals)
	if err != nil || decimals.Uint64() > 18 {
		log.Sugar.Debugf("tx: %s, error: %s, decimals: %s", tx.Hash().Hex(), "decimals invalid", insc.Decimals)
		return nil
	}

	max, err := utils.StringToBigint(insc.Max)
	if err != nil || max.Uint64() < 1 {
		log.Sugar.Debugf("tx: %s, error: %s, max: %s", tx.Hash().Hex(), "max invalid", insc.Max)
		return nil
	}

	lim, err := utils.StringToBigint(insc.Lim)
	if err != nil || lim.Uint64() < 1 {
		log.Sugar.Debugf("tx: %s, error: %s, lim: %s", tx.Hash().Hex(), "lim invalid", insc.Lim)
		return nil
	}

	if max.Cmp(lim) < 0 {
		log.Sugar.Debugf("tx: %s, error: %s, max: %s, lim: %s", tx.Hash().Hex(), "max must gte lim", insc.Max, insc.Lim)
		return nil
	}

	if new(big.Int).Rem(max, lim).Uint64() != 0 {
		log.Sugar.Debugf("tx: %s, error: %s, max: %s, lim: %s", tx.Hash().Hex(), "lim can not divisible by max", insc.Max, insc.Lim)
		return nil
	}

	for _, miner := range insc.Miners {
		if !utils.IsValidERCAddress(miner) {
			log.Sugar.Debugf("tx: %s, error: %s, miner: %s", tx.Hash().Hex(), "invalid miner", miner)
			return nil
		}
	}

	hash := tx.Hash().Hex()
	// add inscription
	inscriptionModel := &dao.InscriptionModel{
		Tick:     insc.Tick,
		TickHash: hash,
		TxIndex:  uint64(index),
		Block:    block.NumberU64(),
		BlockAt:  block.Time(),
		Decimals: uint8(decimals.Uint64()),
		Max:      max.String(),
		Lim:      lim.String(),
		Miners:   strings.Join(insc.Miners, ","),
		Minted:   "0",
		Status:   1,
		Protocol: insc.P,
		DeployBy: strings.ToLower(utils.GetTxFrom(tx).Hex()),
	}
	if inscriptionModel.Id, err = dao.GenSnowflakeID(); err != nil {
		return err
	}
	if err = s.inscriptionDao.Create(db, inscriptionModel); err != nil {
		return err
	}

	miners := mapset.NewSet[string]()
	for _, miner := range insc.Miners {
		miners.Add(miner)
	}
	s.inscriptions[hash] = &inscription{
		Id:       inscriptionModel.Id,
		Max:      max,
		Lim:      lim,
		Minted:   big.NewInt(0),
		Tick:     insc.Tick,
		TickHash: hash,
		Miners:   miners,
	}

	return nil
}

func (s *BscScanService) recap() error {
	return nil
}

func (s *BscScanService) mint(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index int, isPending ...bool) error {
	if len(isPending) > 0 {
		return nil
	}
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
	if insc.Miners.Cardinality() > 0 && !insc.Miners.Contains(strings.ToLower(block.Coinbase().Hex())) {
		log.Sugar.Debugf("tx: %s, error: %s, want: %s, get: %s", tx.Hash().Hex(), "miners", insc.Miners, block.Coinbase().Hex())
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	to := strings.ToLower(tx.To().Hex())
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
		"balance": new(big.Int).Add(balance, amt).String(),
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

	// add record
	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		BlockAt:  block.Time(),
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		TickHash: insc.TickHash,
		From:     from,
		To:       to,
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
	minted := new(big.Int).Add(insc.Minted, amt)
	inscUpdate := map[string]interface{}{
		"minted": minted.String(),
	}
	if minted.Cmp(insc.Max) == 0 {
		inscUpdate["status"] = 2 // mint end
	}
	if err = s.inscriptionDao.Update(db, insc.Id, inscUpdate); err != nil {
		return err
	}

	insc.Minted = minted
	return nil
}

func (s *BscScanService) transfer(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index, opIndex int, isPending ...bool) error {
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

	amt, err := s.transferForFrom(db, block, tx, inscription, index, opIndex, isPending...)
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

func (s *BscScanService) transferForFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index, opIndex int, isPending ...bool) (*big.Int, error) {
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
		OpIndex:  uint64(opIndex),
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

func (s *BscScanService) burn(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48Inscription, index, opIndex int, isPending ...bool) error {
	if block.NumberU64() < 34886548 {
		return nil
	}

	inscription.To = "0x000000000000000000000000000000000000dead"
	return s.transfer(db, block, tx, inscription, index, opIndex, isPending...)
}

func (s *BscScanService) approve() error {
	return nil
}

func (s *BscScanService) transferFrom() error {
	return nil
}

func (s *BscScanService) updateRam(record *dao.AccountRecordsModel, block *types.Block) {
	s.pendingTxs.Lock()
	defer s.pendingTxs.Unlock()

	if s.pendingTxs.TxsHash.Contains(record.TxHash) {
		return
	}
	record.IsPending = true
	changes, err := utils.InputToBNB48Inscription(record.Input, record.Block)
	if err == nil && int(record.OpIndex) < len(changes) {
		record.InputDecode = changes[record.OpIndex]
	}

	s.pendingTxs.TxsInBlock.Add(block.NumberU64())
	s.pendingTxs.TxsHash.Add(record.TxHash)

	s.pendingTxs.Txs[record.TxHash] = record

	if _, ok := s.pendingTxs.TxsByAddr[record.From]; !ok {
		s.pendingTxs.TxsByAddr[record.From] = map[string]types2.RecordsModelByTxHash{
			record.TickHash: {},
		}
	}
	s.pendingTxs.TxsByAddr[record.From][record.TickHash][record.TxHash] = record
	if _, ok := s.pendingTxs.TxsByAddr[record.To]; !ok {
		s.pendingTxs.TxsByAddr[record.To] = map[string]types2.RecordsModelByTxHash{
			record.TickHash: {},
		}
	}
	s.pendingTxs.TxsByAddr[record.To][record.TickHash][record.TxHash] = record

	if _, ok := s.pendingTxs.TxsByTickHash[record.TickHash]; !ok {
		s.pendingTxs.TxsByTickHash[record.TickHash] = types2.RecordsModelByTxHash{
			record.TxHash: {},
		}
	}
	s.pendingTxs.TxsByTickHash[record.TickHash][record.TxHash] = record

}
