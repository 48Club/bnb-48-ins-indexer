package dao

import (
	"bnb-48-ins-indexer/pkg/utils"
	"errors"
	"fmt"
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
	LoadChanges(db *gorm.DB, model *AccountWalletModel, relimit int) error
}

type AccountWalletModel struct {
	Id        uint64                `json:"id,string" gorm:"primaryKey"`
	AccountId uint64                `json:"account_id,string"`
	Address   string                `json:"address"`
	Tick      string                `json:"tick"`
	TickHash  string                `json:"tick_hash"`
	Decimals  uint8                 `json:"decimals" gorm:"-"`
	Balance   string                `json:"balance"`
	Changes   []AccountRecordsModel `json:"changes" gorm:"-"`
	CreateAt  int64                 `json:"create_at"`
	UpdateAt  int64                 `json:"update_at"`
	DeleteAt  int64                 `json:"delete_at"`
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

	if err = db.Table("inscription AS a").Joins("LEFT JOIN account_wallet AS b ON a.tick_hash = b.tick_hash AND b.address = ?", address).Where("a.tick_hash in ?", tickHash).Select("b.`id`", "b.`account_id`", "b.`address`", "b.`balance`", "b.`create_at`", "b.`delete_at`", "b.`update_at`", "a.`decimals`", "a.`tick`", "a.`tick_hash`").Find(&model).Error; err != nil {
		return nil, err
	}

	return model, nil
}
func (h *AccountWalletHandler) SelectByAddress(db *gorm.DB, address string) ([]*AccountWalletModel, error) {
	var (
		model []*AccountWalletModel
		err   error
	)

	if err = db.Table(h.TableName()).Where("address = ?", address).Find(&model).Error; err != nil {
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

	model.CreateAt = time.Now().Unix()
	model.UpdateAt = model.CreateAt

	return db.Table(h.TableName()).Create(model).Error
}

func (h *AccountWalletHandler) UpdateBalance(db *gorm.DB, id uint64, data map[string]interface{}) error {
	var err error

	data["update_at"] = time.Now().Unix()
	if err = db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error; err != nil {
		return err
	}

	return nil
}

func (h *AccountWalletHandler) LoadChanges(db *gorm.DB, model *AccountWalletModel, relimit int) error {
	if relimit >= 20 {
		return nil
	}
	accountRecordsModel := []AccountRecordsModel{}
	db = db.Table((&AccountRecordsHandler{}).TableName())
	addresss := utils.Address2Format(model.Address)
	for k, v := range addresss {
		if k == 0 {
			db.Where("input like ?", fmt.Sprintf("%%%s%%", v))
			continue
		}
		db.Or("input like ?", fmt.Sprintf("%%%s%%", v))
	}
	if len(addresss) == 0 {
		return nil
	}

	errors.Is(db.Error, gorm.ErrRecordNotFound)
	err := db.Limit(20 - relimit).Order("block desc, tx_index desc, op_index desc").Find(&accountRecordsModel).Error
	for _, v := range accountRecordsModel {
		v.InputDecode, _ = utils.InputToBNB48Inscription(v.Input)
		model.Changes = append(model.Changes, v)
	}
	return err
}
