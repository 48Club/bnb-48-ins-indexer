package dao

import (
	"github.com/jwrookie/fans/pkg/utils"
	"gorm.io/gorm"
	"time"
)

type IAccountRecords interface {
	TableName() string
	Create(db *gorm.DB, model *AccountRecordsModel) error
}

type AccountRecordsModel struct {
	Id       uint64 `json:"id,string" gorm:"primaryKey"`
	Block    uint64 `json:"block"`
	BlockAt  uint64 `json:"block_at"`
	TxHash   string `json:"tx_hash"`
	TxIndex  uint64 `json:"tx_index"`
	From     string `json:"from"`
	To       string `json:"to"`
	Input    string `json:"input"`
	Type     uint8  `json:"type"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
	DeleteAt int64  `json:"delete_at"`
}

type AccountRecordsHandler struct {
}

func (h *AccountRecordsHandler) TableName() string {
	return "account_records"
}

func (h *AccountRecordsHandler) Create(db *gorm.DB, model *AccountRecordsModel) error {
	var err error

	// init
	if model.Id == 0 {
		if model.Id, err = utils.GenSnowflakeID(); err != nil {
			return err
		}
	}

	model.CreateAt = time.Now().UnixMilli()
	model.UpdateAt = model.CreateAt

	return db.Table(h.TableName()).Create(model).Error
}
