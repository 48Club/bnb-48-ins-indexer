package controler

import (
	"github.com/gin-gonic/gin"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"github.com/jwrookie/fans/pkg/utils"
	"github.com/jwrookie/fans/service"
)

type RecordController struct {
	recordS *service.RecordService
}

func NewRecordController() *RecordController {
	return &RecordController{
		recordS: service.NewRecordService(),
	}
}

func (c *RecordController) List(ctx *gin.Context) {
	var req bnb48types.CommonListCond
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
