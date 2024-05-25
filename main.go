package main

import (
	"bnb-48-ins-indexer/cmd/api"
	"bnb-48-ins-indexer/cmd/index"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/scripts/upgrade/pr65"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
)

var PendingTxs types.GlobalVariable

func main() {
	go api.Start(&PendingTxs)
	go index.Start(&PendingTxs)
	select {}
}

func init() {
	log.Init("index.log")

	PendingTxs = types.GlobalVariable{
		Txs:           make(types.RecordsModelByTxHash),
		TxsHash:       mapset.NewSet[string](),
		TxsInBlock:    mapset.NewSet[uint64](),
		TxsByTickHash: make(map[string]types.RecordsModelByTxHash),
		TxsByAddr:     make(map[string]map[string]types.RecordsModelByTxHash),
		BlockAt:       common.Big0,
	}

	pr65.Upgrade() // upgrade db, more detail: https://github.com/48Club/bnb-48-ins-indexer/pull/65/files
}
