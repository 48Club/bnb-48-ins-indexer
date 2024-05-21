package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"strconv"
	"strings"
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

func (s *WrapService) Delete(req bnb48types.DeleteWrapReq) error {
	var ids []uint64
	for _, id := range req.Ids {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		ids = append(ids, uint64(i))
	}

	return s.WrapDao.Delete(database.Mysql(), ids, strings.ToLower(req.Hash))
}
