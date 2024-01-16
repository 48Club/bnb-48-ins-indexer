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
	"encoding/json"
	"errors"
	"fmt"
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

const (
	FutureEnableBNForPR61 uint64 = 35_084_848 // more detail: https://github.com/48Club/bnb-48-ins-indexer/pull/61
	FutureEnableBNForPR67 uint64 = 48_484_848 // more detail: https://github.com/48Club/bnb-48-ins-indexer/pull/67
)

type BscScanService struct {
	account        dao.IAccount
	accountRecords dao.IAccountRecords
	accountWallet  dao.IAccountWallet
	inscriptionDao dao.IInscription
	inscriptions   map[string]*inscription // tick-hash : inscription
	conf           config.Config
	pendingTxs     *types2.GlobalVariable
	allowance      dao.IAllowance
}

type inscription struct {
	Id       uint64
	Miners   mapset.Set[string]
	Minters  mapset.Set[string]
	Block    uint64
	Commence uint64
	Max      *big.Int
	Lim      *big.Int
	Minted   *big.Int
	Tick     string
	TickHash string
	DeployBy string
}

var allowRamUpdate = mapset.NewSet[string]()

func NewBscScanService(pendingTxs *types2.GlobalVariable) *BscScanService {
	allowRamUpdate.Append("transferFrom", "transfer")
	return &BscScanService{
		account:        &dao.AccountHandler{},
		accountRecords: &dao.AccountRecordsHandler{},
		accountWallet:  &dao.AccountWalletHandler{},
		inscriptionDao: &dao.InscriptionHandler{},
		allowance:      &dao.AllowanceHandler{},
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
			insc    inscription
			miners  = mapset.NewSet[string]()
			minters = mapset.NewSet[string]()
		)
		insc.Id = ele.Id
		insc.Tick = ele.Tick
		insc.TickHash = ele.TickHash
		insc.DeployBy = ele.DeployBy
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
			miners.Append(strings.Split(ele.Miners, ",")...)
		}

		insc.Miners = miners

		if ele.Block >= FutureEnableBNForPR67 {
			if len(ele.Minters) >= 42 {
				minters.Append(strings.Split(ele.Minters, ",")...)
			}
			insc.Minters = minters
			insc.Commence = ele.Commence
		}

		insc.Block = ele.Block

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
		s.pendingTxs.IndexBloukAt = types2.BlockInfo{
			Number:    targetBN,
			Timestamp: targetBlock.Time(),
		}
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
		if len(isPending) > 0 && isPending[0] && !allowRamUpdate.ContainsOne(data.Op) {
			continue
		}

		switch data.Op {
		case "deploy":
			if err = s.deploy(db, block, tx, data, index); err != nil {
				return err
			}
		case "recap":
			if err = s.recap(db, block, tx, data, index); err != nil {
				return err
			}
		case "mint":
			if err = s.mint(db, block, tx, data, index); err != nil {
				return err
			}
		case "transfer":
			if err = s.transfer(db, block, tx, data, index, opIndex, isPending...); err != nil {
				return err
			}
		case "burn":
			if err = s.burn(db, block, tx, data, index, opIndex); err != nil {
				return err
			}
		case "approve":
			if err = s.approve(db, block, tx, data, index, opIndex); err != nil {
				return err
			}
		case "transferFrom":
			if err = s.transferFrom(db, block, tx, data, index, opIndex, isPending...); err != nil {
				return err
			}
		default:
			log.Sugar.Debugf("tx: %s, error: can not support %s op", tx.Hash().Hex(), data.Op)
		}
	}

	return nil
}

func (s *BscScanService) deploy(db *gorm.DB, block *types.Block, tx *types.Transaction, insc *helper.BNB48InscriptionVerified, index int) error {
	hash := tx.Hash().Hex()
	// add inscription
	inscriptionModel := &dao.InscriptionModel{
		Tick:     insc.Tick,
		TickHash: hash,
		TxIndex:  uint64(index),
		Block:    block.NumberU64(),
		BlockAt:  block.Time(),
		Decimals: uint8(insc.DecimalsV.Uint64()),
		Max:      insc.MaxV.String(),
		Lim:      insc.LimV.String(),
		Miners:   strings.Join(insc.Miners, ","),
		Minted:   insc.ReservesSum.String(),
		Holders:  uint64(len(insc.Reserves)),
		Status:   1,
		Protocol: insc.P,
		DeployBy: strings.ToLower(utils.GetTxFrom(tx).Hex()),
	}

	if inscriptionModel.Block >= FutureEnableBNForPR67 {
		if len(insc.Commence) > 0 {
			commence, err := utils.StringToBigint(insc.Commence)
			if err != nil {
				log.Sugar.Debugf("tx: %s, error: %s, commence: %s", tx.Hash().Hex(), "commence invalid", insc.Commence)
				return nil
			}
			inscriptionModel.Commence = commence.Uint64()
		}
		inscriptionModel.Minters = strings.Join(insc.Minters, ",")
		b, err := json.Marshal(insc.Reserves)
		if err == nil {
			inscriptionModel.Reserves = string(b)
		}

		_ins := &helper.BNB48InscriptionVerified{
			BNB48Inscription: &helper.BNB48Inscription{
				P:        insc.P,
				Op:       "mint",
				TickHash: hash,
			},
		}
		i := 0
		for to, amt := range insc.ReservesV {
			_ins.Amt = amt.String()
			if err := s.createRecord(db, tx, block, _ins, index, i, "0x0000000000000000000000000000000000000000", to, helper.InscriptionStatusMint, false); err != nil {
				return err
			}
			account, err := s.accountCheck(db, to)
			if err != nil {
				return err
			}
			if err = s.accountWallet.Create(db, &dao.AccountWalletModel{
				AccountId: account.Id,
				Tick:      insc.Tick,
				TickHash:  hash,
				Address:   to,
				Balance:   amt.String(),
			}); err != nil {
				return err
			}
			i++
		}
	}

	if err := s.inscriptionDao.Create(db, inscriptionModel); err != nil {
		return err
	}

	miners := mapset.NewSet[string]()
	miners.Append(insc.Miners...)
	minters := mapset.NewSet[string]()
	minters.Append(insc.Minters...)

	s.inscriptions[hash] = &inscription{
		Id:       inscriptionModel.Id,
		Max:      insc.MaxV,
		Lim:      insc.LimV,
		Minted:   insc.ReservesSum,
		Tick:     insc.Tick,
		TickHash: hash,
		Miners:   miners,
		Minters:  minters,
		Block:    inscriptionModel.Block,
		Commence: inscriptionModel.Commence,
	}

	return nil
}

func (s *BscScanService) recap(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index int) error {
	if block.NumberU64() < FutureEnableBNForPR61 {
		return nil
	}

	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "not deploy")
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())

	if from != insc.DeployBy {
		log.Sugar.Debugf("tx: %s, error: %s, deployBy: %s, tx from: %s", tx.Hash().Hex(), "insufficient authority", insc.DeployBy, from)
		return nil
	}

	if inscription.MaxV.Cmp(insc.Max) > 0 {
		log.Sugar.Debugf("tx: %s, error: %s, max: %s", tx.Hash().Hex(), "new target max", insc.Max)
		return nil
	}

	if inscription.MaxV.Cmp(insc.Lim) < 0 {
		log.Sugar.Debugf("tx: %s, error: %s, max: %s, lim: %s", tx.Hash().Hex(), "max must gte lim", insc.Max, insc.Lim)
		return nil
	}

	if new(big.Int).Rem(inscription.MaxV, insc.Lim).Uint64() != 0 {
		log.Sugar.Debugf("tx: %s, error: %s, max: %s, lim: %s", tx.Hash().Hex(), "lim can not divisible by max", insc.Max, insc.Lim)
		return nil
	}

	inscUpdate := map[string]interface{}{
		"max": inscription.Max,
	}
	if insc.Minted.Cmp(inscription.MaxV) >= 0 {
		inscUpdate["status"] = 2 // mint end
	}
	if err := s.inscriptionDao.Update(db, insc.Id, inscUpdate); err != nil {
		return err
	}

	// add record
	if err := s.createRecord(db, tx, block, inscription, index, 0, from, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusRecap); err != nil {
		return err
	}

	insc.Max = inscription.MaxV
	return nil
}

func (s *BscScanService) mint(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index int) error {
	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "not deploy")
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	if insc.Block >= FutureEnableBNForPR67 {
		// commence
		if insc.Commence > 0 && block.NumberU64() < insc.Commence {
			log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "can not mint before commence")
			return nil
		}

		// sender whitelist
		if insc.Minters.Cardinality() > 0 && !insc.Minters.ContainsOne(from) {
			log.Sugar.Debugf("tx: %s, error: %s, want: %s, get: %s", tx.Hash().Hex(), "minters", insc.Minters, from)
			return nil
		}
	}

	// miners
	if insc.Miners.Cardinality() > 0 && !insc.Miners.ContainsOne(strings.ToLower(block.Coinbase().Hex())) {
		log.Sugar.Debugf("tx: %s, error: %s, want: %s, get: %s", tx.Hash().Hex(), "miners", insc.Miners, block.Coinbase().Hex())
		return nil
	}

	// minting ended
	if insc.Minted.Cmp(insc.Max) >= 0 {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "minting ended")
		return nil
	}
	// minting overflow
	if new(big.Int).Add(inscription.AmtV, insc.Minted).Cmp(insc.Max) > 0 {
		log.Sugar.Debugf("tx: %s, error: %s", tx.Hash().Hex(), "minting overflow")
		return nil
	}

	to := strings.ToLower(tx.To().Hex())
	// account
	account, err := s.account.SelectByAddress(db, to)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = &dao.AccountModel{Address: to}
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
		"balance": new(big.Int).Add(balance, inscription.AmtV).String(),
	}
	if err = s.accountWallet.UpdateBalanceByID(db, accountWallet.Id, updates); err != nil {
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
	if err = s.createRecord(db, tx, block, inscription, index, 0, from, to, helper.InscriptionStatusMint); err != nil {
		return err
	}

	// update inscription
	minted := new(big.Int).Add(insc.Minted, inscription.AmtV)
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

func (s *BscScanService) transfer(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int, isPending ...bool) error {
	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s, tick-hash: %s", tx.Hash().Hex(), "not deploy", inscription.TickHash)
		return nil
	}

	amt, err := s.transferForFrom(db, block, tx, inscription, index, opIndex, isPending...)
	if err != nil {
		return err
	}

	if err = s.transferForTo(db, amt, insc, inscription.To, isPending...); err != nil {
		return err
	}

	return nil
}

func (s *BscScanService) transferForFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int, isPending ...bool) (*big.Int, error) {
	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	zero := new(big.Int)

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
	if currentBalance.Cmp(inscription.AmtV) < 0 {
		log.Sugar.Debugf("tx: %s, error: %s, current balance: %s, need balance:%s", tx.Hash().Hex(), "insufficient balance", accountWallet.Balance, inscription.Amt)
		return zero, nil
	}

	// add record for tx from
	balance := new(big.Int).Sub(currentBalance, inscription.AmtV)
	updates := map[string]interface{}{
		"balance": balance.String(),
	}
	if err = s.accountWallet.UpdateBalanceByID(db, accountWallet.Id, updates); err != nil {
		return nil, err
	}
	if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, -1); err != nil {
			return nil, err
		}
	}
	if err = s.createRecord(db, tx, block, inscription, index, opIndex, from, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusTransfer, isPending...); err != nil {
		return nil, err
	}

	return inscription.AmtV, nil
}
func (s *BscScanService) accountCheck(db *gorm.DB, to string) (*dao.AccountModel, error) {
	account, err := s.account.SelectByAddress(db, to)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			account = &dao.AccountModel{Address: to}
			if err = s.account.Create(db, account); err != nil {
				return account, err
			}
		} else {
			return account, err
		}
	}
	return account, err
}

func (s *BscScanService) transferForTo(db *gorm.DB, amt *big.Int, insc *inscription, to string, isPending ...bool) error {
	// account
	account, err := s.accountCheck(db, to)
	if err != nil {
		return err
	}

	// update accountWallet
	newWallet := false
	accountWallet, err := s.accountWallet.SelectByAccountIdTickHash(db, account.Id, insc.TickHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			accountWallet = &dao.AccountWalletModel{AccountId: account.Id, Tick: insc.Tick, TickHash: insc.TickHash, Address: account.Address}
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
	if err = s.accountWallet.UpdateBalanceByID(db, accountWallet.Id, updates); err != nil {
		return err
	}
	if newWallet {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, 1); err != nil {
			return err
		}
	}
	return nil
}

func (s *BscScanService) burn(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int) error {
	if block.NumberU64() < FutureEnableBNForPR61 {
		return nil
	}

	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s, tick-hash: %s", tx.Hash().Hex(), "not deploy", inscription.TickHash)
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())

	// check balance of tx from
	accountWallets, _ := s.accountWallet.SelectByAddressTickHash(db, from, []string{inscription.TickHash})
	if len(accountWallets) == 0 {
		return nil
	}
	accountWallet := accountWallets[0]

	currentBalance, err := utils.StringToBigint(accountWallet.Balance)
	if err != nil {
		return err
	}

	balanceCmp := currentBalance.Cmp(inscription.AmtV)
	if balanceCmp == -1 {
		log.Sugar.Debugf("tx: %s, error: %s, current balance: %s, need balance:%s", tx.Hash().Hex(), "insufficient balance", accountWallet.Balance, inscription.Amt)
		return nil
	}

	// add record for tx from
	balance := common.Big0
	if balanceCmp == 1 {
		balance = new(big.Int).Sub(currentBalance, inscription.AmtV)
	}
	updates := map[string]interface{}{
		"balance": balance.String(),
	}
	if err = s.accountWallet.UpdateBalanceByID(db, accountWallet.Id, updates); err != nil {
		return err
	}
	if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, accountWallet.TickHash, -1); err != nil {
			return err
		}
	}

	if err = s.createRecord(db, tx, block, inscription, index, opIndex, from, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusBurn); err != nil {

		return err
	}

	// update inscription max
	newMax := new(big.Int).Sub(insc.Max, inscription.AmtV)
	inscUpdate := map[string]interface{}{
		"max": newMax.String(),
	}
	if err = s.inscriptionDao.Update(db, insc.Id, inscUpdate); err != nil {
		return err
	}

	insc.Max = newMax
	return nil
}

func (s *BscScanService) approve(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int) error {
	if block.NumberU64() < FutureEnableBNForPR61 {
		return nil
	}

	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s, tick-hash: %s", tx.Hash().Hex(), "not deploy", inscription.TickHash)
		return nil
	}

	owner := strings.ToLower(utils.GetTxFrom(tx).Hex())

	// add record
	if err := s.createRecord(db, tx, block, inscription, index, opIndex, owner, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusApprove); err != nil {
		return err
	}

	// create or update allowance
	return s.allowance.CreateOrUpdate(db, &dao.AllowanceModel{
		Tick:     insc.Tick,
		TickHash: insc.TickHash,
		Owner:    owner,
		Spender:  inscription.Spender,
		Amt:      inscription.AmtV.String(),
		Position: fmt.Sprintf("%d_%d_%d", block.NumberU64(), index, opIndex),
	})
}

func (s *BscScanService) transferFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int, isPending ...bool) error {
	if block.NumberU64() < FutureEnableBNForPR61 {
		return nil
	}

	from := strings.ToLower(utils.GetTxFrom(tx).Hex())
	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		log.Sugar.Debugf("tx: %s, error: %s, tick-hash: %s", tx.Hash().Hex(), "not deploy", inscription.TickHash)
		return nil
	}

	if tmpCmp := inscription.AmtV.Cmp(big.NewInt(0)); tmpCmp == 0 {
		return s.createRecord(db, tx, block, inscription, index, opIndex, from, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusTransferFrom, isPending...)
		// } else if tmpCmp == -1 { // 由于 StringToBigint 会将负数直接抛出错误, 所以这里不需要判断, 注释掉这一部分代码
		// 	log.Sugar.Debugf("tx: %s, error: %s, amt: %s", tx.Hash().Hex(), "invaild amt", "0")
		// 	return nil
	}

	// check balance of owner
	ownerWallets, err := s.accountWallet.SelectByAddressTickHash(db, inscription.From, []string{inscription.TickHash})
	if err != nil {
		return err
	}
	if len(ownerWallets) == 0 {
		return nil
	}

	ownerWallet := ownerWallets[0]

	currentBalance, err := utils.StringToBigint(ownerWallet.Balance)
	if err != nil {
		return err
	}
	currentBalanceCmp := currentBalance.Cmp(inscription.AmtV)
	if currentBalanceCmp < 0 {
		log.Sugar.Debugf("tx: %s, error: %s, current balance: %s, need balance:%s", tx.Hash().Hex(), "insufficient balance", ownerWallet.Balance, inscription.Amt)
		return nil
	}

	// check and allowance
	allowance, err := s.allowance.Select(db, map[string]interface{}{
		"tick_hash": insc.TickHash,
		"owner":     inscription.From,
		"spender":   from,
	})
	if err != nil {
		return utils.Error(err, gorm.ErrRecordNotFound, tx.Hash().Hex(), "allowance not found")
	}
	allowanceAmt := utils.MustStringToBigint(allowance.Amt)
	if allowanceAmt.Cmp(inscription.AmtV) == -1 {
		log.Sugar.Debugf("tx: %s, error: %s, amt: %s, allowance amt: %s", tx.Hash().Hex(), "insufficient amt", inscription.Amt, allowance.Amt)
		return nil
	}

	// update allowance of owner
	allowanceUpdates := map[string]interface{}{
		"amt": new(big.Int).Sub(allowanceAmt, inscription.AmtV).String(),
	}
	if err = s.allowance.Update(db, allowance.Id, allowanceUpdates); err != nil {
		return err
	}

	// update balance of owner
	balance := common.Big0
	if currentBalanceCmp == 1 {
		balance = new(big.Int).Sub(currentBalance, inscription.AmtV)
	}
	updates := map[string]interface{}{
		"balance": balance.String(),
	}
	if err = s.accountWallet.UpdateBalanceByAddressTickHash(db, ownerWallet.Address, ownerWallet.TickHash, updates); err != nil {
		return err
	}
	if balance.Cmp(common.Big0) == 0 {
		if err = s.inscriptionDao.UpdateHolders(db, ownerWallet.TickHash, -1); err != nil {
			return err
		}
	}

	// add record for tx from

	if err = s.createRecord(db, tx, block, inscription, index, opIndex, from, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusTransferFrom, isPending...); err != nil {
		return err
	}

	return s.transferForTo(db, inscription.AmtV, insc, inscription.To)
}

func (s *BscScanService) createRecord(db *gorm.DB, tx *types.Transaction, block *types.Block, inscription *helper.BNB48InscriptionVerified, index, opIndex int, from, to string, txType helper.AccountRecordsType, isPending ...bool) error {
	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		OpIndex:  uint64(opIndex),
		TickHash: inscription.TickHash,
		BlockAt:  block.Time(),
		From:     from,
		To:       to,
		Input:    hexutil.Encode(tx.Data()),
		Type:     txType,
		OpJson: func() string {
			b, _ := json.Marshal(inscription.BNB48Inscription)
			return string(b)
		}(),
	}
	if len(isPending) > 0 && isPending[0] {
		s.updateRam(record, block)
	}
	return s.accountRecords.Create(db, record)
}

func (s *BscScanService) updateRam(record *dao.AccountRecordsModel, block *types.Block) {
	// todo: 添加 opIndex 判断
	s.pendingTxs.Lock()
	defer s.pendingTxs.Unlock()

	if s.pendingTxs.TxsHash.ContainsOne(record.TxHash) {
		return
	}
	record.IsPending = true
	changes, err := utils.InputToBNB48Inscription(record.Input, record.Block)
	if err == nil && int(record.OpIndex) < len(changes) {
		record.InputDecode = changes[record.OpIndex].BNB48Inscription
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
