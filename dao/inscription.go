package dao

import (
	"github.com/jwrookie/fans/pkg/utils"
	"gorm.io/gorm"
	"time"
)

type IInscription interface {
	TableName() string
	Create(db *gorm.DB, model *InscriptionModel) error
	Find(db *gorm.DB) ([]*InscriptionModel, error)
}

type InscriptionModel struct {
	Id       uint64 `json:"id,string" gorm:"primaryKey"`
	Tick     string `json:"tick"`
	TickHash string `json:"tick_hash"`
	TxIndex  uint64 `json:"tx_index"`
	Block    uint64 `json:"block"`
	BlockAt  uint64 `json:"block_at"`
	Decimal  uint8  `json:"decimal"`
	Max      string `json:"max"`
	Lim      string `json:"lim"`
	Miners   string `json:"miners"`
	Minted   string `json:"minted"`
	CreateAt int64  `json:"create_at"`
	UpdateAt int64  `json:"update_at"`
	DeleteAt int64  `json:"delete_at"`
}

type InscriptionHandler struct {
}

func (h *InscriptionHandler) TableName() string {
	return "inscription"
}

func (h *InscriptionHandler) Find(db *gorm.DB) ([]*InscriptionModel, error) {
	var (
		datas []*InscriptionModel
		err   error
	)

	db = db.Where("delete_at = ?", 0)

	if err = db.Table(h.TableName()).Find(&datas).Error; err != nil {
		return nil, err
	}

	return datas, nil
}

func (h *InscriptionHandler) Create(db *gorm.DB, model *InscriptionModel) error {
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
