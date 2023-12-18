package dao

import (
	"time"

	"gorm.io/gorm"
)

type IAccount interface {
	TableName() string
	Create(db *gorm.DB, model *AccountModel) error
	SelectByAddress(db *gorm.DB, address string) (*AccountModel, error)
}

type AccountModel struct {
	Id       uint64 `json:"id,string" gorm:"primaryKey"`
	Address  string `json:"address"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
	DeleteAt int64  `json:"delete_at"`
}

type AccountHandler struct {
}

func (h *AccountHandler) TableName() string {
	return "account"
}

func (h *AccountHandler) Create(db *gorm.DB, model *AccountModel) error {
	var err error

	// init
	if model.Id == 0 {
		if model.Id, err = GenSnowflakeID(); err != nil {
			return err
		}
	}

	model.CreateAt = time.Now().Unix()
	model.UpdateAt = model.CreateAt

	return db.Table(h.TableName()).Create(model).Error
}

func (h *AccountHandler) SelectByAddress(db *gorm.DB, address string) (*AccountModel, error) {
	var (
		model AccountModel
		err   error
	)

	if err = db.Table(h.TableName()).Where("address = ?", address).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}
