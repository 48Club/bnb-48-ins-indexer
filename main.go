package main

import (
	"bnb-48-ins-indexer/cmd/api"
	"bnb-48-ins-indexer/cmd/index"
	"bnb-48-ins-indexer/dao"
)

var PendingTxs []dao.AccountRecordsModel

func main() {
	api.Start(&PendingTxs)
	index.Start(&PendingTxs)
}
