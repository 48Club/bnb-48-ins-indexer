package controler

import (
	"github.com/gin-gonic/gin"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"github.com/jwrookie/fans/pkg/utils"
	"github.com/jwrookie/fans/service"
)

type AccountController struct {
	accountS *service.AccountService
}

func NewAccountController() *AccountController {
	return &AccountController{
		accountS: service.NewAccountService(),
	}
}

func (c *AccountController) List(ctx *gin.Context) {
	var req bnb48types.ListAccountWalletReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.accountS.List(req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}
