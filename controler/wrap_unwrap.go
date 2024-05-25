package controler

import (
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"
	"github.com/gin-gonic/gin"
)

type WrapController struct {
	wrap *service.WrapService
}

func NewWrapController() *WrapController {
	return &WrapController{
		wrap: service.NewWrapService(),
	}
}

func (c *WrapController) List(ctx *gin.Context) {
	var req bnb48types.ListWrapReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.wrap.List(req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, bnb48types.ListWrapRsp{
		List: res,
	})
}

func (c *WrapController) Delete(ctx *gin.Context) {
	var req bnb48types.DeleteWrapReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	err := c.wrap.Delete(req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, nil)
}
