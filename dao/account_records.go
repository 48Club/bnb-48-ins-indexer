package dao

import (
	"time"

	"bnb-48-ins-indexer/pkg/helper"

	"gorm.io/gorm"
)

type IAccountRecords interface {
	TableName() string
	Create(db *gorm.DB, model *AccountRecordsModel) error
	Find(db *gorm.DB) ([]*AccountRecordsModel, error)
	Count(db *gorm.DB) (int64, error)
}

type AccountRecordsModel struct {
	Id          uint64                   `json:"id,string" gorm:"primaryKey"`
	Block       uint64                   `json:"block"`
	BlockAt     uint64                   `json:"block_at"` // timestamp in second
	TxHash      string                   `json:"tx_hash"`
	OpIndex     uint64                   `json:"op_index"`
	TxIndex     uint64                   `json:"tx_index"`
	TickHash    string                   `json:"tick_hash"`
	From        string                   `json:"from"`
	To          string                   `json:"to"`
	Input       string                   `json:"input"`
	InputDecode *helper.BNB48Inscription `json:"input_decode" gorm:"-"`
	Type        uint8                    `json:"type"`
	CreateAt    int64                    `json:"create_at"`
	UpdateAt    int64                    `json:"update_at"`
	DeleteAt    int64                    `json:"delete_at"`
}

type AccountRecordsHandler struct{}

func (h *AccountRecordsHandler) TableName() string {
	return "account_records"
}

func (h *AccountRecordsHandler) Create(db *gorm.DB, model *AccountRecordsModel) error {
	var err error

	// init
	if model.Id == 0 {
		if model.Id, err = GenSnowflakeID(); err != nil {
			return err
		}
	}

	model.CreateAt = time.Now().UnixMilli()
	model.UpdateAt = model.CreateAt

	return db.Table(h.TableName()).Create(model).Error
}

func (h *AccountRecordsHandler) Find(db *gorm.DB) ([]*AccountRecordsModel, error) {
	var (
		datas []*AccountRecordsModel
		err   error
	)

	db = db.Where("delete_at = 0")

	if err = db.Table(h.TableName()).Find(&datas).Error; err != nil {
		return nil, err
	}

	return datas, nil
}

func (h *AccountRecordsHandler) Count(db *gorm.DB) (int64, error) {
	var (
		res int64
		err error
	)

	db = db.Where("delete_at = 0")

	if err = db.Table(h.TableName()).Count(&res).Error; err != nil {
		return 0, err
	}

	return res, nil
}
