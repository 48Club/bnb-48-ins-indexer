package dao

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type IAllowance interface {
	TableName() string
	Create(db *gorm.DB, model *AllowanceModel) error
	Select(db *gorm.DB, filter map[string]interface{}) (*AllowanceModel, error)
	Update(db *gorm.DB, id uint64, data map[string]interface{}) error
	CreateOrUpdate(db *gorm.DB, model *AllowanceModel) error
}

type AllowanceModel struct {
	Id       uint64 `json:"id,string" gorm:"primaryKey"`
	Tick     string `json:"tick"`
	TickHash string `json:"tick_hash"`
	Owner    string `json:"owner"`
	Spender  string `json:"spender"`
	Amt      string `json:"amt"`
	Position string `json:"position"`
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
	var model AllowanceModel

	db = db.Where("delete_at = 0")

	for k, v := range filter {
		if v == nil || v == "" {
			continue
		}
		db = db.Where(fmt.Sprintf("%s = ?", k), v)
	}

	tx := db.Table(h.TableName()).Select(&model)
	return &model, tx.Error
}

func (h *AllowanceHandler) Update(db *gorm.DB, id uint64, data map[string]interface{}) error {

	data["update_at"] = time.Now().Unix()

	return db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error
}

func (h *AllowanceHandler) CreateOrUpdate(db *gorm.DB, model *AllowanceModel) error {
	updates := map[string]interface{}{
		"amt":       model.Amt,
		"position":  model.Position,
		"update_at": time.Now().Unix(),
	}
	tx := db.Table(h.TableName()).Where("owner = ? AND spender = ? AND tick_hash = ?", model.Owner, model.Spender, model.TickHash).Updates(updates)
	if tx.RowsAffected == 1 {
		return nil
	}
	return h.Create(db, model)
}
