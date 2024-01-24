package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"fmt"
	"os"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestRamDataInBulk(t *testing.T) {
	pendingTxs := types.GlobalVariable{
		Txs:           make(types.RecordsModelByTxHash),
		TxsHash:       mapset.NewSet[string](),
		TxsInBlock:    mapset.NewSet[uint64](),
		TxsByTickHash: make(map[string]types.RecordsModelByTxHash),
		TxsByAddr:     make(map[string]map[string]types.RecordsModelByTxHash),
		BlockAt:       common.Big0,
	}
	ser := NewBscScanService(&pendingTxs)
	txHash := "0xba6e736afc6587c05d93fe22ddc6a53e380698ad233b05974881c961eb9bb938"
	fileBytes, err := os.ReadFile(fmt.Sprintf("../test/%s", txHash))
	assert.NoError(t, err)
	strHex := string(fileBytes)
	inss, err := utils.InputToBNB48Inscription(strHex, 34_862_697)
	assert.NoError(t, err)

	addresss := []string{"0xaacc290a1a4c89f5d7bc29913122f5982916de48"}
	for opIndex, ins := range inss {
		ser.updateRam(&dao.AccountRecordsModel{
			Input:    strHex,
			TxHash:   txHash,
			TickHash: ins.TickHash,
			Block:    34_862_697,
			From:     addresss[0],
			To:       addresss[0],
			OpIndex:  uint64(opIndex),
		}, 34_862_697)
		addresss = append(addresss, ins.To)
	}

	for i, address := range addresss {
		changes := []dao.AccountRecordsModel{}
		if _txsByAddr, ok := pendingTxs.TxsByAddr[address]; ok {
			if _txsByTickHash, ok := _txsByAddr[inss[0].TickHash]; ok {
				for _, tx := range _txsByTickHash {
					changes = append(changes, *tx)
				}
			}
		}

		if i == 0 {
			assert.Equal(t, 41, len(changes))
		} else {
			assert.Equal(t, 1, len(changes))
		}
	}

}
