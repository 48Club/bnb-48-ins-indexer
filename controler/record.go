package controler

import (
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

	res, err := c.recordS.List(req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}
