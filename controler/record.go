package controler

import (
	"bnb-48-ins-indexer/dao"
	"bnb-48-ins-indexer/pkg/types"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"

	"github.com/gin-gonic/gin"
)

type RecordController struct {
	recordS    *service.RecordService
	pendingTxs *types.GlobalVariable
}

func NewRecordController(pendingTxs *types.GlobalVariable) *RecordController {
	return &RecordController{
		recordS:    service.NewRecordService(),
		pendingTxs: pendingTxs,
	}
}

func (c *RecordController) List(ctx *gin.Context) {
	var req bnb48types.ListRecordReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	resList := []*dao.AccountRecordsModel{}
	ramTx := c.pendingTxs.Txs.ToSlice()
	ramTxLen := int64(len(ramTx))
	if ramTxLen > 0 {
		if ramTxLen >= int64(req.PageSize)*(req.Page+1) {
			resList = ramTx[ramTxLen-int64(req.PageSize)*(req.Page) : ramTxLen-int64(req.PageSize)*(req.Page+1)]
		} else if ramTxLen > int64(req.PageSize)*req.Page {
			resList = ramTx[ramTxLen-int64(req.PageSize)*req.Page:]
		}
	}
	res, err := c.recordS.List(req)

	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}
	if lrl := len(resList); lrl < int(req.PageSize) {
		// 将 res.List 中的数据补充到 resList , 数量不能超过 req.PageSize
		for i := 0; i < int(req.PageSize)-lrl; i++ {
			if i >= len(res.List) {
				break
			}
			resList = append(resList, res.List[i])
		}
	}
	res.List = resList

	utils.SuccessResponse(ctx, res)
}
