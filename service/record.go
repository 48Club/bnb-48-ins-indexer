package service

import (
	"github.com/jwrookie/fans/dao"
	"github.com/jwrookie/fans/pkg/database"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"gorm.io/gorm"
)

type RecordService struct {
	recordDao dao.IAccountRecords
}

func NewRecordService() *RecordService {
	return &RecordService{
		recordDao: &dao.AccountRecordsHandler{},
	}
}

func (s *RecordService) List(req bnb48types.CommonListCond) (*bnb48types.ListRecordRsp, error) {
	db := database.Mysql()
	var res []*dao.AccountRecordsModel
	var count int64
	if err := db.Transaction(func(tx *gorm.DB) error {
		countTx := tx.Session(&gorm.Session{Context: tx.Statement.Context})
		if req.PageSize > 0 {
			tx = tx.Limit(int(req.PageSize))
		}
		tx = tx.Offset(int(req.Page) * int(req.PageSize))
		var err error
		res, err = s.recordDao.Find(tx)
		if err != nil {
			return err
		}
		count, err = s.recordDao.Count(countTx)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	resp := &bnb48types.ListRecordRsp{
		CommonListRsp: bnb48types.CommonListRsp{
			Count:    count,
			Page:     uint64(req.Page),
			PageSize: uint8(req.PageSize),
		},
		List: res,
	}
	return resp, nil
}
