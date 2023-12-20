package main

import (
	"bnb-48-ins-indexer/cmd/api"
	"bnb-48-ins-indexer/cmd/index"
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/types"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
)

var PendingTxs types.GlobalVariable

func main() {
	api.Start(&PendingTxs)
	index.Start(&PendingTxs)
}

func init() {
	PendingTxs = types.GlobalVariable{
		Txs:           mapset.NewSet[*dao.AccountRecordsModel](),
		TxsInBlock:    mapset.NewSet[uint64](),
		TxsByTickHash: map[string]mapset.Set[*dao.AccountRecordsModel]{},
		TxsByAddr:     map[common.Address]map[string]mapset.Set[dao.AccountRecordsModel]{},
		BlockAt:       common.Big0,
	}
}
