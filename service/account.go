package service

import (
	"github.com/jwrookie/fans/dao"
	"github.com/jwrookie/fans/pkg/database"
	bnb48types "github.com/jwrookie/fans/pkg/types"
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
	var res *dao.AccountWalletModel
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error

		tx = tx.Where("tick_hash = ?, address = ?", req.TickHash, req.Address)

		res, err = s.walletDao.SelectByAddressTickHash(tx, req.Address, req.TickHash)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	resp := &bnb48types.AccountBalanceRsp{AccountWalletModel: res}
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
