package service

import "github.com/jwrookie/fans/dao"

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
