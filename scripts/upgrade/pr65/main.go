package pr65

import (
	_ "bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/utils"
	"encoding/json"
)

func Upgrade() {
	database.NewMysql()
	db := database.Mysql()

	accountRecords := &dao.AccountRecordsHandler{}
	datas := []dao.AccountRecordsModel{}
	tx := db.Table(accountRecords.TableName()).Where("`op_json` IS NULL").Find(&datas)
	if tx.Error != nil {
		// if op_json column not exist, pls merge pr65 sql change first
		panic(tx.Error)
	}
	if tx.RowsAffected == 0 {
		// no data need to upgrade
		return
	}

	for _, ele := range datas {
		input, err := utils.InputToBNB48Inscription(ele.Input, ele.Block)
		if err != nil {
			panic(err)
		}
		if int(ele.OpIndex) >= len(input) {
			panic("op_index out of range")
		}
		b, err := json.Marshal(input[ele.OpIndex])
		if err != nil {
			panic(err)
		}
		if err := db.Table("account_records").Where("`id` = ?", ele.Id).Update("op_json", string(b)); err != nil {
			panic(err)
		}
	}
	log.Sugar.Info("upgrade pr65 success")
}
