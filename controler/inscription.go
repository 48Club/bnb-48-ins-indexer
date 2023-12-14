package controler

import (
	"github.com/gin-gonic/gin"
	bnb48types "github.com/jwrookie/fans/pkg/types"
	"github.com/jwrookie/fans/pkg/utils"
	"github.com/jwrookie/fans/service"
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
	var req bnb48types.CommonListCond
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
