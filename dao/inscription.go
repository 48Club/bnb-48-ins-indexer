package dao

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type IInscription interface {
	TableName() string
	Create(db *gorm.DB, model *InscriptionModel) error
	Find(db *gorm.DB) ([]*InscriptionModel, error)
	Count(db *gorm.DB) (int64, error)
	UpdateHolders(db *gorm.DB, tick string, delta int64) error
	Update(db *gorm.DB, id uint64, data map[string]interface{}) error
}

type InscriptionModel struct {
	Id          uint64 `json:"id,string" gorm:"primaryKey"`
	Tick        string `json:"tick"`
	TickHash    string `json:"tick_hash"`
	TxIndex     uint64 `json:"tx_index"`
	Block       uint64 `json:"block"`
	BlockAt     uint64 `json:"block_at"`
	Decimals    uint8  `json:"decimals"`
	Max         string `json:"max"`
	Lim         string `json:"lim"`
	Miners      string `json:"miners"`
	Minted      string `json:"minted"`
	Status      uint64 `json:"status"`
	Protocol    string `json:"protocol"`
	Holders     uint64 `json:"holders"`
	Information string `json:"information"`
	DeployBy    string `json:"deploy_by"`
	CreateAt    int64  `json:"create_at"`
	UpdateAt    int64  `json:"update_at"`
	DeleteAt    int64  `json:"delete_at"`
}

type InscriptionHandler struct{}

func (h *InscriptionHandler) TableName() string {
	return "inscription"
}

func (h *InscriptionHandler) Count(db *gorm.DB) (int64, error) {
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

func (h *InscriptionHandler) Find(db *gorm.DB) ([]*InscriptionModel, error) {
	var (
		datas []*InscriptionModel
		err   error
	)

	db = db.Where("delete_at = 0")

	if err = db.Table(h.TableName()).Find(&datas).Error; err != nil {
		return nil, err
	}

	return datas, nil
}

func (h *InscriptionHandler) Create(db *gorm.DB, model *InscriptionModel) error {
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

func (h *InscriptionHandler) UpdateHolders(db *gorm.DB, tick string, delta int64) error {
	// var err error
	var res []*InscriptionModel
	db = db.Where("tick = ? ", tick)
	if err := db.Table(h.TableName()).Find(&res).Error; err != nil {
		return err
	}
	if len(res) == 0 {
		return fmt.Errorf("tick %s invalid", tick)
	}
	model := res[0]
	holders := int64(model.Holders) + delta
	model.Holders += uint64(holders)
	if err := db.Table(h.TableName()).Save(model).Error; err != nil {
		return err
	}
	return nil
}

func (h *InscriptionHandler) Update(db *gorm.DB, id uint64, data map[string]interface{}) error {
	var err error

	data["update_at"] = time.Now().UnixMilli()
	if err = db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error; err != nil {
		return err
	}

	return nil
}
