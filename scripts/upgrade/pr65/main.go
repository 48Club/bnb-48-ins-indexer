package pr65

import (
	_ "bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/utils"
	"encoding/json"
	"fmt"
)

func Upgrade() {
	database.NewMysql()
	db := database.Mysql()

	accountRecords := &dao.AccountRecordsHandler{}
	datas := []dao.AccountRecordsModel{}
	tx := db.Table(accountRecords.TableName()).Where("`op_json` IS NULL").Find(&datas)
	if err := tx.Error; err != nil {
		panic(fmt.Sprintf("if op_json column not exist, pls merge pr65 sql change first, err: %s", err.Error()))
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
		if err := db.Table(accountRecords.TableName()).Where("`id` = ?", ele.Id).Update("`op_json`", string(b)).Error; err != nil {
			panic(err)
		}
	}

	log.Sugar.Info("upgrade pr65 success")
}
