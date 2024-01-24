package types

import (
	"bnb-48-ins-indexer/dao"
	"fmt"
	"math/big"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
)

func BuildTxsHashKeyWithOpIndex(txHash string, opIndex uint64) string {
	return fmt.Sprintf("%s-%d", txHash, opIndex)
}

type GlobalVariable struct {
	Txs           RecordsModelByTxHash
	TxsHash       mapset.Set[string]
	TxsByTickHash map[string]RecordsModelByTxHash
	TxsInBlock    mapset.Set[uint64]
	TxsByAddr     map[string]map[string]RecordsModelByTxHash
	BlockAt       *big.Int
	IndexBloukAt  BlockInfo
	mu            sync.Mutex
}

func (g *GlobalVariable) UpdateTxsByAddr(address, tickHash, txHash string, record *dao.AccountRecordsModel) {
	if _, ok := g.TxsByAddr[address]; !ok {
		g.TxsByAddr[address] = map[string]RecordsModelByTxHash{
			tickHash: {},
		}
	}
	g.TxsByAddr[address][tickHash][txHash] = record
}

func (g *GlobalVariable) UpdateTxsByTickHash(tickHash, txHash string, record *dao.AccountRecordsModel) {
	if _, ok := g.TxsByTickHash[tickHash]; !ok {
		g.TxsByTickHash[tickHash] = RecordsModelByTxHash{}
	}
	g.TxsByTickHash[tickHash][txHash] = record
}

type BlockInfo struct {
	Number    *big.Int `json:"number"`
	Timestamp uint64   `json:"timestamp"`
}

type RecordsModelByTxHash map[string]*dao.AccountRecordsModel

func (g *GlobalVariable) Lock() {
	g.mu.Lock()
}

func (g *GlobalVariable) Unlock() {
	g.mu.Unlock()
}
