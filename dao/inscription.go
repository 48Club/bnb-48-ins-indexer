package dao

import (
	"errors"
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
	Id       uint64 `json:"id,string" gorm:"primaryKey"`
	Tick     string `json:"tick"`
	TickHash string `json:"tick_hash"`
	TxIndex  uint64 `json:"tx_index"`
	Block    uint64 `json:"block"`
	Commence uint64 `json:"commence"`
	BlockAt  uint64 `json:"block_at"`
	Decimals uint8  `json:"decimals"`
	Max      string `json:"max"`
	Lim      string `json:"lim"`
	Miners   string `json:"miners"`
	Minters  string `json:"minters"`
	Reserves string `json:"reserves"`
	Minted   string `json:"minted"`
	Status   uint64 `json:"status"`
	Protocol string `json:"protocol"`
	Holders  uint64 `json:"holders"`
	DeployBy string `json:"deploy_by"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
	DeleteAt int64  `json:"delete_at"`
}

type InscriptionHandler struct{}

func (h *InscriptionHandler) TableName() string {
	return "inscription"
}

func (h *InscriptionHandler) Count(db *gorm.DB) (int64, error) {
	var res int64

	tx := db.Table(h.TableName()).Where("delete_at = 0").Count(&res)

	return res, tx.Error
}

func (h *InscriptionHandler) Find(db *gorm.DB) ([]*InscriptionModel, error) {
	var datas []*InscriptionModel

	tx := db.Table(h.TableName()).Where("delete_at = 0").Find(&datas)

	return datas, tx.Error
}

func (h *InscriptionHandler) Create(db *gorm.DB, model *InscriptionModel) error {
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

func (h *InscriptionHandler) UpdateHolders(db *gorm.DB, tick_hash string, delta int64) error {
	var res *InscriptionModel
	db = db.Where("tick_hash = ? ", tick_hash)
	if err := db.Table(h.TableName()).First(&res).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("tick %s invalid", tick_hash)
	} else if err != nil {
		return err
	}

	return db.Table(h.TableName()).Update("holders", int64(res.Holders)+delta).Error
}

func (h *InscriptionHandler) Update(db *gorm.DB, id uint64, data map[string]interface{}) error {

	data["update_at"] = time.Now().Unix()

	return db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error
}
