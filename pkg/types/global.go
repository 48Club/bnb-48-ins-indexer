package types

import (
	"bnb-48-ins-indexer/dao"

	"github.com/ethereum/go-ethereum/common"
)

type GlobalVariable struct {
	Txs        []dao.AccountRecordsModel
	TxsInBlock map[uint64][]dao.AccountRecordsModel
	TxsByAddr  map[common.Address][]dao.AccountRecordsModel
}
