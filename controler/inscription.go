package controler

import (
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/service"
)

type InscriptionController struct {
	inscriptionS *service.InscriptionService
}

func (c *InscriptionController) List(ctx *gin.Context) {

}
