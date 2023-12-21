package types

import (
	"bnb-48-ins-indexer/dao"
	"math/big"

	mapset "github.com/deckarep/golang-set/v2"
)

type GlobalVariable struct {
	Txs           RecordsModelByTxHash
	TxsHash       mapset.Set[string]
	TxsByTickHash map[string]RecordsModelByTxHash
	TxsInBlock    mapset.Set[uint64]
	TxsByAddr     map[string]map[string]RecordsModelByTxHash
	BlockAt       *big.Int
}

type RecordsModelByTxHash map[string]*dao.AccountRecordsModel
