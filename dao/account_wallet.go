package dao

import (
	"github.com/jwrookie/fans/pkg/utils"
	"gorm.io/gorm"
	"time"
)

type IAccountWallet interface {
	TableName() string
	Create(db *gorm.DB, model *AccountWalletModel) error
	UpdateBalance(db *gorm.DB, id uint64, data map[string]interface{}) error
	SelectByAccountIdTickHash(db *gorm.DB, accountId uint64, tickHash string) (*AccountWalletModel, error)
}

type AccountWalletModel struct {
	Id        uint64 `json:"id,string" gorm:"primaryKey"`
	AccountId uint64 `json:"account_id,string"`
	Tick      string `json:"tick"`
	TickHash  string `json:"tick_hash"`
	Balance   string `json:"balance"`
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
	DeleteAt  int64  `json:"delete_at"`
}

type AccountWalletHandler struct {
}

func (h *AccountWalletHandler) TableName() string {
	return "account_wallet"
}

func (h *AccountWalletHandler) SelectByAccountIdTickHash(db *gorm.DB, accountId uint64, tickHash string) (*AccountWalletModel, error) {
	var (
		model AccountWalletModel
		err   error
	)

	if err = db.Table(h.TableName()).Where("account_id = ? and tick_hash = ?", accountId, tickHash).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (h *AccountWalletHandler) Create(db *gorm.DB, model *AccountWalletModel) error {
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

func (h *AccountWalletHandler) UpdateBalance(db *gorm.DB, id uint64, data map[string]interface{}) error {
	var err error

	data["update_at"] = time.Now().UnixMilli()
	if err = db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error; err != nil {
		return err
	}

	return nil
}
