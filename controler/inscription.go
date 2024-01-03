package controler

import (
	"bnb-48-ins-indexer/pkg/types"
	bnb48types "bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/pkg/utils"
	"bnb-48-ins-indexer/service"

	"github.com/gin-gonic/gin"
)

type InscriptionController struct {
	inscriptionS *service.InscriptionService
	pendingTxs   *types.GlobalVariable
}

func NewInscriptionController(pendingTxs *types.GlobalVariable) *InscriptionController {
	return &InscriptionController{
		inscriptionS: service.NewInscriptionService(),
		pendingTxs:   pendingTxs,
	}
}

func (c *InscriptionController) List(ctx *gin.Context) {
	var req bnb48types.ListInscriptionWalletReq
	if err := ctx.ShouldBind(&req); err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	res, err := c.inscriptionS.List(req, c.pendingTxs.IndexBloukAt)
	if err != nil {
		utils.FailResponse(ctx, err.Error())
		return
	}

	utils.SuccessResponse(ctx, res)
}
