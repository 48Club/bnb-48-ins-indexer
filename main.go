package main

import (
	"bnb-48-ins-indexer/cmd/api"
	"bnb-48-ins-indexer/cmd/index"
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/types"

	"github.com/ethereum/go-ethereum/common"
)

var PendingTxs types.GlobalVariable

func main() {
	api.Start(&PendingTxs)
	index.Start(&PendingTxs)
}

func init() {
	PendingTxs = types.GlobalVariable{
		Txs:        []dao.AccountRecordsModel{},
		TxsInBlock: map[uint64][]dao.AccountRecordsModel{},
		TxsByAddr:  map[common.Address][]dao.AccountRecordsModel{},
	}
}
