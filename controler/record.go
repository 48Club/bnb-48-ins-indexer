package controler

import (
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/service"
)

type RecordController struct {
	recordS *service.RecordService
}

func (c *RecordController) List(ctx *gin.Context) {

}
