package controler

import (
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/service"
)

type AccountController struct {
	accountS *service.AccountService
}

func (c *AccountController) List(ctx *gin.Context) {

}
