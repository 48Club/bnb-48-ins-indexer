package service

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/database"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"

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

func (s *RecordService) List(req bnb48types.ListRecordReq) (*bnb48types.ListRecordRsp, error) {
	db := database.Mysql()
	var res []*dao.AccountRecordsModel
	var count int64
	if err := db.Transaction(func(tx *gorm.DB) error {

		tx = tx.Order("`block` desc, `tx_index` desc")

		if req.TickHash != "" {
			tx = tx.Where("`tick_hash` = ?", req.TickHash)
		}
		var err error
		count, err = s.recordDao.Count(tx)
		if err != nil {
			return err
		}
		if req.PageSize > 0 {
			tx = tx.Limit(int(req.PageSize))
		}
		tx = tx.Offset(int(req.Page) * int(req.PageSize))

		res, err = s.recordDao.Find(tx)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	for k, v := range res {
		v.InputDecode, _ = utils.InputToBNB48Inscription(v.Input, v.Block)
		res[k] = v
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

func (s *RecordService) Get(req bnb48types.GetRecordReq) (*bnb48types.GetRecordRsp, error) {
	lists, err := s.recordDao.FindByTxHash(database.Mysql(), req.TxHash)
	if err != nil {
		return nil, err
	}

	for _, ele := range lists {
		ele.InputDecode, _ = utils.InputToBNB48Inscription(ele.Input, ele.Block)
	}

	return &bnb48types.GetRecordRsp{
		List: lists,
	}, nil
}
