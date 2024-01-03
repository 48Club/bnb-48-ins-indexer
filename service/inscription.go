package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	bnb48types "bnb-48-ins-indexer/pkg/types"

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

type Status uint64

const (
	InProgress Status = 1
	Completed  Status = 2
)

func (s *InscriptionService) List(req bnb48types.ListInscriptionWalletReq, bn bnb48types.BlockInfo) (*bnb48types.ListInscriptionRsp, error) {
	db := database.Mysql()
	var res []*dao.InscriptionModel
	var count int64
	if err := db.Transaction(func(tx *gorm.DB) error {

		tx = tx.Order("`holders` DESC, `create_at` DESC")

		if req.Protocol != "" {
			tx = tx.Where("protocol = ?", req.Protocol)
		}
		if req.Status > 0 {
			tx = tx.Where("status = ?", req.Status)
		}
		if req.TickHash != "" {
			tx = tx.Where("tick_hash = ?", req.TickHash)
		}
		if req.Tick != "" {
			tx = tx.Where("tick = ?", req.Tick)
		}
		if req.DeployBy != "" {
			tx = tx.Where("deploy_by = ?", req.DeployBy)
		}

		var err error
		count, err = s.inscriptionDao.Count(tx)
		if err != nil {
			return err
		}
		if req.PageSize > 0 {
			tx = tx.Limit(int(req.PageSize))
		}
		tx = tx.Offset(int(req.Page) * int(req.PageSize))

		res, err = s.inscriptionDao.Find(tx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	resp := &bnb48types.ListInscriptionRsp{
		CommonListRsp: bnb48types.BuildResponseInfo(count, req.Page, req.PageSize, bn),
		List:          res,
	}
	return resp, nil
}
