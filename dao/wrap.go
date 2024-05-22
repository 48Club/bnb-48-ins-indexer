package dao

import (
	"gorm.io/gorm"
	"time"
)

type IWrap interface {
	TableName() string
	Create(db *gorm.DB, model *WrapModel) error
	Delete(db *gorm.DB, ids []uint64, txHash string) error
	List(db *gorm.DB, limit, t uint64) ([]WrapModel, error)
	FindByIds(db *gorm.DB, ids []uint64) ([]WrapModel, error)
}

type WrapModel struct {
	Id         uint64 `json:"id,string" gorm:"primaryKey"`
	TickHash   string `json:"tick_hash"`
	TxHash     string `json:"tx_hash"`
	To         string `json:"to"`
	Amt        string `json:"amt"`
	WrapTxHash string `json:"wrap_tx_hash"`
	Type       uint8  `json:"type"`
	CreateAt   int64  `json:"create_at"`
	UpdateAt   int64  `json:"update_at"`
	DeleteAt   int64  `json:"delete_at"`
}

type WrapHandler struct {
}

func (h *WrapHandler) TableName() string {
	return "wrap"
}

func (h *WrapHandler) Create(db *gorm.DB, model *WrapModel) error {
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

func (h *WrapHandler) Delete(db *gorm.DB, ids []uint64, txHash string) error {
	updates := map[string]interface{}{
		"wrap_tx_hash": txHash,
		"delete_at":    time.Now().Unix(),
	}

	return db.Table(h.TableName()).Where("id in ?", ids).UpdateColumns(updates).Error
}

func (h *WrapHandler) List(db *gorm.DB, limit, t uint64) ([]WrapModel, error) {
	var datas []WrapModel

	tx := db.Table(h.TableName()).Where("delete_at = 0 and type = ?", t).Order("create_at asc").Limit(int(limit)).Find(&datas)

	return datas, tx.Error
}

func (h *WrapHandler) FindByIds(db *gorm.DB, ids []uint64) ([]WrapModel, error) {
	var datas []WrapModel

	tx := db.Table(h.TableName()).Where("delete_at = 0 and id in ?", ids).Find(&datas)

	return datas, tx.Error
}
