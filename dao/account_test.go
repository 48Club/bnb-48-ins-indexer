package dao

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestFindAccount(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("error0")
	}
	defer conn.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      conn,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatal("error1")
	}
	rows := sqlmock.NewRows([]string{
		"id", "block", "block_at", "tx_hash", "tx_index", "from", "to", "input", "type", "create_at", "update_at", "delete_at",
	}).AddRow(1, 1, 1, "abc", 1, "from", "to", "input", 0, 1, 1, 0)
	mock.ExpectQuery("^SELECT (.+) FROM `account_records`").WillReturnRows(rows)
	h := AccountRecordsHandler{}
	records, err := h.Find(db)
	if err != nil {
		t.Fatal("error2")
	}
	if len(records) != 1 {
		t.Fatal("error3")
	}
}
