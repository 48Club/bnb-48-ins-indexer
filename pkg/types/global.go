package types

import (
	"bnb-48-ins-indexer/dao"
	"math/big"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
)

type GlobalVariable struct {
	Txs           mapset.Set[*dao.AccountRecordsModel]
	TxsByTickHash map[string]mapset.Set[*dao.AccountRecordsModel]
	TxsInBlock    mapset.Set[uint64]
	TxsByAddr     map[common.Address]map[string]mapset.Set[dao.AccountRecordsModel]
	BlockAt       *big.Int
}
