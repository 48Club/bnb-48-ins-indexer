package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/config"
	"github.com/jwrookie/fans/controler"
)

func NewBotRoute() *gin.Engine {
	r := gin.Default()
	conf := config.GetConfig()
	bnb48 := r.Group(conf.App.RoutePrefix)

	var (
		account     = controler.NewAccountController()
		record      = controler.NewRecordController()
		inscription = controler.NewInscriptionController()
	)

	v1 := bnb48.Group("/v1")
	{
		v1.POST("/balance/list", account.List)
		v1.POST("/record/list", record.List)
		v1.POST("/inscription/list", inscription.List)
	}

	return r
}