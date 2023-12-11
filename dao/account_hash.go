package dao

import (
	"github.com/jwrookie/fans/pkg/utils"
	"gorm.io/gorm"
	"time"
)

type IAccountHash interface {
	TableName() string
	Create(db *gorm.DB, model *AccountHashModel) error
	Select(db *gorm.DB, filter *AccountHashModel) (*AccountHashModel, error)
	Update(db *gorm.DB, id uint64, data map[string]interface{}) error
}

type AccountHashModel struct {
	Id        uint64 `json:"id,string" gorm:"primaryKey"`
	AccountId uint64 `json:"account_id"`
	MintHash  string `json:"mint_hash"`
	State     uint8  `json:"state"`
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
	DeleteAt  int64  `json:"delete_at"`
}

type AccountHashHandler struct {
}

func (h *AccountHashHandler) TableName() string {
	return "account_hash"
}

func (h *AccountHashHandler) Create(db *gorm.DB, model *AccountHashModel) error {
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

func (h *AccountHashHandler) Select(db *gorm.DB, filter *AccountHashModel) (*AccountHashModel, error) {
	var (
		model AccountHashModel
		err   error
	)

	db = db.Table(h.TableName())
	if filter.AccountId != 0 {
		db = db.Where("account_id = ?", filter.AccountId)
	}
	if filter.MintHash != "" {
		db = db.Where("mint_hash = ?", filter.MintHash)
	}
	if filter.State != 0 {
		db = db.Where("state = ?", filter.State)
	}

	db = db.Where("delete_at = ?", 0)

	if err = db.First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (h *AccountHashHandler) Update(db *gorm.DB, id uint64, data map[string]interface{}) error {
	var err error

	data["update_at"] = time.Now().UnixMilli()
	if err = db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error; err != nil {
		return err
	}

	return nil
}
