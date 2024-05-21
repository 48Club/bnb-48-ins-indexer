package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	bnb48types "bnb-48-ins-indexer/pkg/types"
)

type WrapService struct {
	WrapDao dao.IWrap
}

func NewWrapService() *WrapService {
	return &WrapService{
		WrapDao: &dao.WrapHandler{},
	}
}

func (s *WrapService) List(req bnb48types.ListWrapReq) ([]dao.WrapModel, error) {
	return s.WrapDao.List(database.Mysql(), 50, req.Type)
}
