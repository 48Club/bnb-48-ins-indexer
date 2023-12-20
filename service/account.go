package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	bnb48types "bnb-48-ins-indexer/pkg/types"

	"gorm.io/gorm"
)

type AccountService struct {
	accountDao dao.IAccount
	walletDao  dao.IAccountWallet
}

func NewAccountService() *AccountService {
	return &AccountService{
		accountDao: &dao.AccountHandler{},
		walletDao:  &dao.AccountWalletHandler{},
	}
}

func (s *AccountService) Balance(req bnb48types.AccountBalanceReq) (*bnb48types.AccountBalanceRsp, error) {
	db := database.Mysql()
	var res []*dao.AccountWalletModel
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		if len(req.TickHash) == 0 {
			res, err = s.walletDao.SelectByAddress(tx, req.Address)
		} else {
			res, err = s.walletDao.SelectByAddressTickHash(tx, req.Address, req.TickHash)
		}

		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	resp := &bnb48types.AccountBalanceRsp{Wallet: res}
	return resp, nil
}

func (s *AccountService) List(req bnb48types.ListAccountWalletReq) (*bnb48types.ListAccountWalletRsp, error) {
	db := database.Mysql()
	var (
		res   []*dao.AccountWalletModel
		count int64
	)
	if err := db.Transaction(func(tx *gorm.DB) error {
		countTx := tx.Session(&gorm.Session{Context: tx.Statement.Context})

		tx = tx.Order("CAST(`balance` as UNSIGNED) DESC").Where("CAST(`balance` as UNSIGNED) > 0")
		if req.PageSize > 0 {
			tx = tx.Limit(int(req.PageSize))
		}
		tx = tx.Offset(int(req.Page) * int(req.PageSize))
		var err error

		res, err = s.walletDao.FindByTickHash(tx, req.TickHash)
		if err != nil {
			return err
		}

		count, err = s.walletDao.Count(countTx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	resp := &bnb48types.ListAccountWalletRsp{
		CommonListRsp: bnb48types.CommonListRsp{
			Count:    count,
			Page:     uint64(req.Page),
			PageSize: uint8(req.PageSize),
		},
		List: res,
	}
	return resp, nil
}

func (s *AccountService) GetInscription(insTickHashs []string, inss *[]*dao.InscriptionModel) error {
	return database.Mysql().Table("inscription").Where("tick_hash in ?", insTickHashs).Find(&inss).Error
}
