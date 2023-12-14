package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/dao"
)

type RecordService struct {
	recordDao dao.IAccountRecords
}

func NewRecordService() *RecordService {
	return &RecordService{
		recordDao: &dao.AccountRecordsHandler{},
	}
}

func (c *RecordService) List(ctx *gin.Context) ([]*dao.AccountRecordsModel, error) {

}
