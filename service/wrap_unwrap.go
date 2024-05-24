package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/global"
	"bnb-48-ins-indexer/pkg/helper"
	"bnb-48-ins-indexer/pkg/log"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type WrapService struct {
	*BscScanService
	wrapDao dao.IWrap
}

func NewWrapService() *WrapService {
	for {
		if defaultBscScanService == nil {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		break
	}

	return &WrapService{
		BscScanService: defaultBscScanService,
		wrapDao:        &dao.WrapHandler{},
	}
}

func (s *WrapService) List(req bnb48types.ListWrapReq) ([]dao.WrapModel, error) {
	return s.wrapDao.List(database.Mysql(), 50, req.Type)
}

func (s *WrapService) Delete(req bnb48types.DeleteWrapReq) error {
	var ids []uint64
	for _, id := range req.Ids {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		ids = append(ids, uint64(i))
	}

	rs, err := global.BscClient.TransactionReceipt(context.TODO(), common.HexToHash(req.Hash))
	if err != nil {
		return err
	}
	if rs.Status != 1 {
		return nil
	}

	tx := database.Mysql().Begin()
	defer tx.Rollback()

	if err = s.delete(tx, rs, ids, rs.BlockNumber, rs.TransactionIndex); err != nil {
		return err
	}

	if err = s.wrapDao.Delete(tx, ids, strings.ToLower(req.Hash)); err != nil {
		return err
	}
	return tx.Commit().Error
}

func (s *WrapService) delete(tx *gorm.DB, rs *types.Receipt, ids []uint64, blockNumber *big.Int, index uint) error {
	models, err := s.wrapDao.FindByIds(database.Mysql(), ids)
	if err != nil {
		return err
	}

	if len(models) != len(ids) {
		return errors.New("len(models) != len(ids)")
	}

	uniT := models[0].Type
	for _, model := range models {
		if uniT != model.Type {
			return errors.New("unique required for type")
		}
	}

	uniTickHash := models[0].TickHash
	for _, model := range models {
		if uniTickHash != model.TickHash {
			return errors.New("unique required for TickHash")
		}
	}

	switch uniT {
	case 1:
		return s.deleteForUnWrap(rs, models)
	case 2:
		return s.deleteForWrap(tx, models, rs.TxHash.Hex(), blockNumber, index)
	}

	return nil
}

func (s *WrapService) deleteForWrap(tx *gorm.DB, models []dao.WrapModel, txHash string, blockNumber *big.Int, index uint) error {
	trans, _, err := global.BscClient.TransactionByHash(context.TODO(), common.HexToHash(txHash))
	if err != nil {
		return err
	}

	block, err := global.BscClient.BlockByNumber(context.TODO(), blockNumber)
	if err != nil {
		return err
	}

	if len(trans.Data()) < 4 || !bytes.Equal(trans.Data()[:4], []byte{94, 247, 241, 217}) {
		return errors.New("tx data error")
	}

	var txData string
	r, err := utils.Unpack([]string{"string"}, trans.Data()[4:])
	if err != nil {
		return err
	}

	switch r[0].(type) {
	case string:
		txData = r[0].(string)
	default:
		return errors.New("data parse error")
	}

	datas, err := utils.InputToBNB48Inscription([]byte(txData), blockNumber.Uint64())
	if err != nil {
		return err
	}

	transDataMap := make(map[string]string)
	for _, data := range datas {
		if data == nil {
			return errors.New("data is nil")
		}

		if data.P != "bnb-48" {
			return errors.New("p != bnb-48")
		}

		if data.Op != "transfer" {
			return errors.New("op != transfer")
		}

		if data.TickHash != models[0].TickHash {
			return errors.New("TickHash not eq")
		}

		transDataMap[strings.ToLower(data.To)] += data.Amt
	}

	modelsDataMap := make(map[string]string)
	// check address and amt
	for _, model := range models {
		modelsDataMap[model.To] += model.Amt
	}

	if err = s.checkModels(modelsDataMap, transDataMap); err != nil {
		return err
	}

	if _, err := s.inscriptionDao.Lock(tx); err != nil {
		return err
	}

	for opIndex, data := range datas {
		if err = s.transfer(tx, block, trans, data, int(index), opIndex, hexutil.Encode([]byte(txData))); err != nil {
			return err
		}
	}

	return nil
}

func (s *WrapService) transfer(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int, input string) error {
	insc, ok := s.inscriptions[inscription.TickHash]
	// not deploy
	if !ok {
		return errors.New("not deploy")
	}

	amt, err := s.transferForFrom(db, block, tx, inscription, index, opIndex, input)
	if err != nil {
		return err
	}

	if err = s.transferForTo(db, amt, insc, inscription.To); err != nil {
		return err
	}

	return nil
}

func (s *WrapService) transferForFrom(db *gorm.DB, block *types.Block, tx *types.Transaction, inscription *helper.BNB48InscriptionVerified, index, opIndex int, input string) (*big.Int, error) {
	from := strings.ToLower(utils.GetTxFrom(tx).Hex())

	// sub balance of tx from
	fromAccount, err := s.account.SelectByAddress(db, from)
	if err != nil {
		return nil, err
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
		return nil, errors.New("insufficient balance")
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
	if err = s.createRecord(db, tx, block, inscription, index, opIndex, from, strings.ToLower(tx.To().Hex()), helper.InscriptionStatusTransfer, input); err != nil {
		return nil, err
	}

	return inscription.AmtV, nil
}

func (s *WrapService) createRecord(db *gorm.DB, tx *types.Transaction, block *types.Block, inscription *helper.BNB48InscriptionVerified, index, opIndex int, from, to string, txType helper.AccountRecordsType, input string, isPending ...bool) error {
	record := &dao.AccountRecordsModel{
		Block:    block.NumberU64(),
		TxHash:   tx.Hash().Hex(),
		TxIndex:  uint64(index),
		OpIndex:  uint64(opIndex),
		TickHash: inscription.TickHash,
		BlockAt:  block.Time(),
		From:     from,
		To:       to,
		Input:    input,
		Type:     txType,
		OpJson: func() string {
			b, _ := json.Marshal(inscription.BNB48Inscription)
			return string(b)
		}(),
	}
	if len(isPending) > 0 && isPending[0] {
		s.updateRam(record, block.NumberU64())
	}
	return s.accountRecords.Create(db, record)
}

func (s *WrapService) deleteForUnWrap(rs *types.Receipt, models []dao.WrapModel) error {
	uintTy, err := abi.NewType("uint256", "uint256", nil)
	if err != nil {
		return err
	}

	arguments := abi.Arguments{
		{
			Type: uintTy,
		},
	}

	transDataMap := make(map[string]string)

	for _, event := range rs.Logs {
		var data []interface{}
		if !strings.EqualFold(event.Topics[0].Hex(), "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef") { // transfer
			continue
		}

		if data, err = arguments.UnpackValues(event.Data); err != nil {
			return err
		}
		if len(data) == 0 {
			return errors.New("data invalid")
		}

		to := strings.ToLower(common.BytesToAddress(event.Topics[2].Bytes()).Hex())
		transDataMap[to] += data[0].(*big.Int).String()
	}

	// check address and amt
	modelsDataMap := make(map[string]string)
	for _, model := range models {
		modelsDataMap[model.To] += model.Amt
	}

	return s.checkModels(modelsDataMap, transDataMap)
}

func (s *WrapService) checkModels(models, transData map[string]string) error {
	log.Sugar.Info(models)
	log.Sugar.Info(transData)
	if len(models) != len(transData) {
		return errors.New("len not eq")
	}

	for address, amt := range models {
		_amt, ok := transData[address]
		if !ok {
			return errors.New("check amt error")
		}
		if _amt != amt {
			return errors.New("amt not eq")
		}
	}

	return nil
}
