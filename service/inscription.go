package service

import (
	"github.com/jwrookie/fans/dao"
	"github.com/jwrookie/fans/pkg/database"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"gorm.io/gorm"
)

type InscriptionService struct {
	inscriptionDao dao.IInscription
}

func NewInscriptionService() *InscriptionService {
	return &InscriptionService{
		inscriptionDao: &dao.InscriptionHandler{},
	}
}

func (s *InscriptionService) List(req *bnb48types.CommonListCond) (*bnb48types.ListInscriptionRsp, error) {
	db := database.Mysql()
	var res []*dao.InscriptionModel
	var count int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		countTx := tx.Session(&gorm.Session{Context: tx.Statement.Context})
		if req.PageSize > 0 {
			tx = tx.Limit(int(req.PageSize))
		}
		tx = tx.Offset(int(req.Page) * int(req.PageSize))
		var err error
		res, err = s.inscriptionDao.Find(tx)
		if err != nil {
			return err
		}
		count, err = s.inscriptionDao.Count(countTx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	resp := &bnb48types.ListInscriptionRsp{
		CommonListRsp: bnb48types.CommonListRsp{
			Count:    count,
			Page:     uint64(req.Page),
			PageSize: uint8(req.PageSize),
		},
		List: res,
	}
	return resp, nil
}
