package index

import (
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/service"
)

func Start(pendingTxs *types.GlobalVariable) {
	log.Init("index.log")
	database.NewMysql()

	bsc := service.NewBscScanService(pendingTxs)

	if err := bsc.Scan(); err != nil {
		panic(err)
	}
}
