package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/config"
)

func NewBotRoute() *gin.Engine {
	r := gin.Default()
	conf := config.GetConfig()
	bnb48 := r.Group(conf.App.RoutePrefix)

	v1 := bnb48.Group("/v1")
	v1.GET("/balance/list")
	v1.GET("/record/list")
	v1.GET("/inscription/list")

	return r
}
