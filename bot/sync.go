package main

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/utils"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type user struct {
	txHash    common.Hash
	Validated *big.Int
	Returned  *big.Int
}

var (
	users          map[common.Address]*user = map[common.Address]*user{}
	block          uint64                   = 0
	totalSum       *big.Int                 = big.NewInt(0)
	totalReturnSum *big.Int                 = big.NewInt(0)
	totalValidSum  *big.Int                 = big.NewInt(0)
	maxSupply      *big.Int                 = big.NewInt(20_000_000_000_000)
)

const (
	beginBlock uint64 = 37_000_000
	endBlock   uint64 = 37_131_034
)

func sync() {
	var datas []dao.AccountRecordsModel
	tx := mySql.Table("`account_records`").Where("`tick_hash` = '0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2' AND `op_json_to` = '0x9567443394a3a611A6335Bab0e64f7F5E0cD83fd' AND `op_json_op` = 'transfer' AND `block` BETWEEN ? AND ?", block, endBlock).Order("`block` ASC, `tx_index` ASC, `op_index` ASC").Find(&datas)
	if tx.Error != nil {
		panic(tx.Error)
	}
	for _, data := range datas {
		_ = json.Unmarshal([]byte(data.OpJson), &data.InputDecode)
		f := common.HexToAddress(data.From)
		if data.OpJsonFrom != "" {
			f = common.HexToAddress(data.OpJsonFrom)
		}
		amt := utils.MustStringToBigint(data.InputDecode.Amt)
		u, ok := users[f]
		if ok {
			u.Returned.Add(u.Returned, amt)
			totalReturnSum.Add(totalReturnSum, amt)
		} else {
			tmp_user := user{
				txHash:    common.HexToHash(data.TxHash),
				Validated: big.NewInt(10_000_000_000),
				Returned:  big.NewInt(0),
			}
			if cmp := amt.Cmp(tmp_user.Validated); cmp == 1 {
				tmp_user.Returned.Sub(amt, tmp_user.Validated)
			} else if cmp == -1 {
				tmp_user.Validated = amt
			}

			tmpTotalValidSum := big.NewInt(0).Add(totalValidSum, tmp_user.Validated)
			if tmpTotalValidSum.Cmp(maxSupply) == 1 {
				if cmp := totalValidSum.Cmp(maxSupply); cmp == -1 {
					tmp_user.Returned.Add(tmp_user.Returned, big.NewInt(0).Sub(tmpTotalValidSum, maxSupply))
					tmp_user.Validated.Sub(tmp_user.Validated, tmp_user.Returned)
				} else if cmp >= 0 {
					tmp_user.Returned = amt
					tmp_user.Validated = big.NewInt(0)
				}
			}

			totalReturnSum.Add(totalReturnSum, tmp_user.Returned)
			totalValidSum.Add(totalValidSum, tmp_user.Validated)
			users[f] = &tmp_user
		}
		totalSum.Add(totalSum, amt)
	}
}
