package test

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/types"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestIfInWhere(t *testing.T) {
	/*
		need sql like this:
		SELECT * FROM `account_records` WHERE block >=35033698 AND IF(`block` = 35033698, tx_index >=54 AND op_index >= 0, true) ORDER BY `block` desc, `tx_index` desc LIMIT 240,30;

		note: run this test, you need to start mysql service and set env MYSQL_ROOT_PASSWORD to your mysql root password
	*/
	db, err := gorm.Open(mysql.Open(fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/mysql", os.Getenv("MYSQL_ROOT_PASSWORD"))))

	assert.NoError(t, err)

	req := types.ListRecordReq{
		BlockNumber: 35033698,
		TxIndex:     54,
		OpIndex:     0,
		CommonListCond: types.CommonListCond{
			Page:     8,
			PageSize: 30,
		},
	}

	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		tx = tx.Order("`block` desc, `tx_index` desc")
		if req.TickHash != "" {
			tx = tx.Where("`tick_hash` = ?", req.TickHash)
		}
		if req.BlockNumber != 0 {
			// 根据区块号查询
			tx = tx.Where("`block` >= ?", req.BlockNumber).Where("IF(`block` = ?, tx_index >=? AND op_index >= ?, true)", req.BlockNumber, req.TxIndex, req.OpIndex)
		}
		if req.PageSize > 0 {
			tx = tx.Limit(int(req.PageSize))
		}
		tx = tx.Offset(int(req.Page) * int(req.PageSize))
		tx.Table("account_records").Where("delete_at = 0").Find(&[]dao.AccountRecordsModel{})
		return tx
	})
	wantSQL := "SELECT * FROM `account_records` WHERE `block` >= 35033698 AND (IF(`block` = 35033698, tx_index >=54 AND op_index >= 0, true)) AND delete_at = 0 ORDER BY `block` desc, `tx_index` desc LIMIT 30 OFFSET 240"
	assert.Equal(t, wantSQL, sqlStr)
}
