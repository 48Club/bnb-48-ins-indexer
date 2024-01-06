package dao

import (
	"bnb-48-ins-indexer/pkg/utils"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type IAccountWallet interface {
	TableName() string
	Create(db *gorm.DB, model *AccountWalletModel) error
	UpdateBalanceByID(db *gorm.DB, id uint64, data map[string]interface{}) error
	UpdateBalanceByAddressTickHash(db *gorm.DB, address, tickHash string, data map[string]interface{}) error
	SelectByAccountIdTickHash(db *gorm.DB, accountId uint64, tickHash string) (*AccountWalletModel, error)
	SelectByAddressTickHash(db *gorm.DB, address string, tickHash []string) ([]*AccountWalletModel, error)
	SelectByAddress(db *gorm.DB, address string) ([]*AccountWalletModel, error)
	Find(db *gorm.DB) ([]*AccountWalletModel, error)
	Count(db *gorm.DB) (int64, error)
	LoadChanges(db *gorm.DB, model *AccountWalletModel, relimit int) error
	ORMSelectColumn() []string
}

type AccountWalletModel struct {
	Id        uint64                `json:"id,string" gorm:"primaryKey"`
	AccountId uint64                `json:"account_id,string"`
	Address   string                `json:"address"`
	Tick      string                `json:"tick"`
	TickHash  string                `json:"tick_hash"`
	Decimals  uint8                 `json:"decimals" gorm:"->"`
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
	var res int64

	tx := db.Table(h.TableName()).Where("delete_at = 0").Count(&res)

	return res, tx.Error
}

func (h *AccountWalletHandler) Find(db *gorm.DB) ([]*AccountWalletModel, error) {
	var datas []*AccountWalletModel

	tx := db.Table(h.TableName()).Where("delete_at = 0").Find(&datas)

	return datas, tx.Error
}

func (h *AccountWalletHandler) SelectByAccountIdTickHash(db *gorm.DB, accountId uint64, tickHash string) (*AccountWalletModel, error) {
	var model AccountWalletModel

	tx := db.Table(h.TableName()).Where("delete_at = 0").Where("account_id = ? and tick_hash = ?", accountId, tickHash).First(&model)

	return &model, tx.Error
}

func (h *AccountWalletHandler) ORMSelectColumn() []string {
	return []string{"`account_wallet`.`id`", "`account_wallet`.`account_id`", "`account_wallet`.`address`", "`account_wallet`.`balance`", "`account_wallet`.`create_at`", "`account_wallet`.`delete_at`", "`account_wallet`.`update_at`", "`inscription`.`decimals`", "`inscription`.`tick`", "`inscription`.`tick_hash`"}
}

func (h *AccountWalletHandler) SelectByAddressTickHash(db *gorm.DB, address string, tickHash []string) ([]*AccountWalletModel, error) {
	var model []*AccountWalletModel

	tx := db.Table("`inscription`").Joins("LEFT JOIN `account_wallet` ON `account_wallet`.`tick_hash` = `inscription`.`tick_hash` AND `account_wallet`.`address` = ?", address).Where("`inscription`.`tick_hash` in ?", tickHash).Select(h.ORMSelectColumn()).Find(&model)

	return model, tx.Error
}
func (h *AccountWalletHandler) SelectByAddress(db *gorm.DB, address string) ([]*AccountWalletModel, error) {
	var model []*AccountWalletModel

	tx := db.Table("`account_wallet`").Joins("RIGHT JOIN `inscription` ON `inscription`.`tick_hash` = `account_wallet`.`tick_hash`").Where("`account_wallet`.`address` = ?", address).Select(h.ORMSelectColumn()).Find(&model)

	return model, tx.Error
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

func (h *AccountWalletHandler) UpdateBalanceByID(db *gorm.DB, id uint64, data map[string]interface{}) error {
	data["update_at"] = time.Now().Unix()

	return db.Table(h.TableName()).Where("id = ?", id).UpdateColumns(data).Error
}
func (h *AccountWalletHandler) UpdateBalanceByAddressTickHash(db *gorm.DB, address, tickHash string, data map[string]interface{}) error {
	data["update_at"] = time.Now().Unix()

	return db.Table(h.TableName()).Where("address = ? and tick_hash = ?", address, tickHash).UpdateColumns(data).Error
}

func (h *AccountWalletHandler) LoadChanges(db *gorm.DB, model *AccountWalletModel, relimit int) error {
	if relimit >= 20 {
		return nil
	}
	accountRecordsModel := []AccountRecordsModel{}
	db = db.Table((&AccountRecordsHandler{}).TableName())
	addresss := utils.Address2Format(model.Address)
	db.Where(
		func(as []string, tx *gorm.DB) *gorm.DB {
			for _, v := range as {
				tx.Or("input like ?", fmt.Sprintf("%%%s%%", v))
			}
			return tx
		}(addresss, db).Or("`from` = ?", model.Address),
	).Where("delete_at = 0").Where("`tick_hash` = ?", model.TickHash)
	if len(addresss) == 0 {
		return nil
	}

	err := db.Limit(20 - relimit).Order("block desc, tx_index desc, op_index desc").Find(&accountRecordsModel).Error
	for _, v := range accountRecordsModel {
		changes, err := utils.InputToBNB48Inscription(v.Input, v.Block)
		if err != nil || int(v.OpIndex) >= len(changes) {
			continue
		}
		v.InputDecode = changes[v.OpIndex]
		model.Changes = append(model.Changes, v)
	}
	return err
}
