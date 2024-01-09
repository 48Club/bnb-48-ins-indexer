package test

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/types"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestIfInWhere(t *testing.T) {
	// need sql like this:
	// SELECT * FROM `account_records` WHERE block >=35033698 AND IF(`block` = 35033698, tx_index >=54 AND op_index >= 0, true) ORDER BY `block` desc, `tx_index` desc LIMIT 240,30;
	db, _ := gorm.Open(mysql.Open("root:4C1C909E-712B-4DAF-997C-DBF1B0998443@tcp(127.0.0.1:3306)/bnb48_inscription?charset=utf8mb4"))
	// assert.NoError(t, err)
	req := types.ListRecordReq{
		BlockNumber: 35033698,
		TxIndex:     54,
		OpIndex:     0,
		CommonListCond: types.CommonListCond{
			Page:     8,
			PageSize: 30,
		},
	}
	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {

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

	t.Log(sql) // output: SELECT * FROM `account_records` WHERE `block` >= 35033698 AND (IF(`block` = 35033698, tx_index >=54 AND op_index >= 0, true)) AND delete_at = 0 ORDER BY `block` desc, `tx_index` desc LIMIT 30 OFFSET 240;
}
