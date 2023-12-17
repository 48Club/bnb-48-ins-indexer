package dao

import (
	"time"

	"gorm.io/gorm"
)

type IAccountWallet interface {
	TableName() string
	Create(db *gorm.DB, model *AccountWalletModel) error
	UpdateBalance(db *gorm.DB, id uint64, data map[string]interface{}) error
	SelectByAccountIdTickHash(db *gorm.DB, accountId uint64, tickHash string) (*AccountWalletModel, error)
	SelectByAddressTickHash(db *gorm.DB, address string, tickHash []string) ([]*AccountWalletModel, error)
	SelectByAddress(db *gorm.DB, address string) ([]*AccountWalletModel, error)
	FindByTickHash(db *gorm.DB, tickHash string) ([]*AccountWalletModel, error)
	Count(db *gorm.DB) (int64, error)
}

type AccountWalletModel struct {
	Id        uint64 `json:"id,string" gorm:"primaryKey"`
	AccountId uint64 `json:"account_id,string"`
	Address   string `json:"address"`
	Tick      string `json:"tick"`
	TickHash  string `json:"tick_hash"`
	Balance   string `json:"balance"`
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
	DeleteAt  int64  `json:"delete_at"`
}

type AccountWalletHandler struct{}

func (h *AccountWalletHandler) TableName() string {
	return "account_wallet"
}

func (h *AccountWalletHandler) Count(db *gorm.DB) (int64, error) {
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

func (h *AccountWalletHandler) FindByTickHash(db *gorm.DB, tickHash string) ([]*AccountWalletModel, error) {
	var (
		datas []*AccountWalletModel
		err   error
	)

	db = db.Where("delete_at = 0 and tick_hash = ?", tickHash)

	if err = db.Table(h.TableName()).Find(&datas).Error; err != nil {
		return nil, err
	}

	return datas, nil
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

func (h *AccountWalletHandler) SelectByAddressTickHash(db *gorm.DB, address string, tickHash []string) ([]*AccountWalletModel, error) {
	var (
		model []*AccountWalletModel
		err   error
	)

	if err = db.Table(h.TableName()).Where("address = ?", address).Where("tick_hash in ?", tickHash).Find(&model).Error; err != nil {
		return nil, err
	}

	return model, nil
}

func (h *AccountWalletHandler) SelectByAddress(db *gorm.DB, address string) ([]*AccountWalletModel, error) {
	var (
		model []*AccountWalletModel
		err   error
	)

	if err = db.Table(h.TableName()).Where("address = ? ", address).Find(&model).Error; err != nil {
		return nil, err
	}

	return model, nil
}

func (h *AccountWalletHandler) Create(db *gorm.DB, model *AccountWalletModel) error {
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

func (h *AccountWalletHandler) UpdateBalance(db *gorm.DB, id uint64, data map[string]interface{}) error {
	var err error

	data["update_at"] = time.Now().UnixMilli()
	if err = db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error; err != nil {
		return err
	}

	return nil
}
