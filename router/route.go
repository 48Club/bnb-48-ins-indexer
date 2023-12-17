package router

import (
	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/controler"

	"github.com/gin-gonic/gin"
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
		v1.POST("/account/balance", account.Balance)
	}

	return r
}
