package controler

import (
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"

	"github.com/gin-gonic/gin"
)

type InscriptionController struct {
	inscriptionS *service.InscriptionService
}

func NewInscriptionController() *InscriptionController {
	return &InscriptionController{
		inscriptionS: service.NewInscriptionService(),
	}
}

func (c *InscriptionController) List(ctx *gin.Context) {
	var req bnb48types.ListInscriptionWalletReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.inscriptionS.List(&req)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}
