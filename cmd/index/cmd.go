package index

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/service"
)

func Start(pendingTxs *[]dao.AccountRecordsModel) {
	log.Init("index.log")
	database.NewMysql()

	bsc := service.NewBscScanService()

	if err := bsc.Scan(); err != nil {
		panic(err)
	}
}
