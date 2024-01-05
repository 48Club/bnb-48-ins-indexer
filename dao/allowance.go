package dao

import (
	"gorm.io/gorm"
	"time"
)

type IAllowance interface {
	TableName() string
	Create(db *gorm.DB, model *AllowanceModel) error
	Select(db *gorm.DB, filter map[string]interface{}) (*AllowanceModel, error)
	Update(db *gorm.DB, id uint64, data map[string]interface{}) error
}

type AllowanceModel struct {
	Id       uint64 `json:"id,string" gorm:"primaryKey"`
	Tick     string `json:"tick"`
	TickHash string `json:"tick_hash"`
	Owner    string `json:"owner"`
	Spender  string `json:"spender"`
	Amt      string `json:"amt"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
	DeleteAt int64  `json:"delete_at"`
}

type AllowanceHandler struct {
}

func (h *AllowanceHandler) TableName() string {
	return "allowance"
}

func (h *AllowanceHandler) Create(db *gorm.DB, model *AllowanceModel) error {
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

func (h *AllowanceHandler) Select(db *gorm.DB, filter map[string]interface{}) (*AllowanceModel, error) {
	var (
		model AllowanceModel
		err   error
	)

	db = db.Where("delete_at = 0")

	if filter["owner"] != nil {
		db = db.Where("owner = ?", filter["owner"])
	}

	if filter["spender"] != nil {
		db = db.Where("spender = ?", filter["spender"])
	}

	if filter["tick_hash"] != nil {
		db = db.Where("tick_hash = ?", filter["tick_hash"])
	}

	if err = db.Table(h.TableName()).Select(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (h *AllowanceHandler) Update(db *gorm.DB, id uint64, data map[string]interface{}) error {
	var err error

	data["update_at"] = time.Now().Unix()
	if err = db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error; err != nil {
		return err
	}

	return nil
}
