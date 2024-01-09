package router

import (
	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/controler"
	"bnb-48-ins-indexer/pkg/types"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

const (
	BSCPerBlockTime = 3 * time.Second
)

func NewBotRoute(pendingTxs *types.GlobalVariable) *gin.Engine {
	r := gin.Default()
	conf := config.GetConfig()
	bnb48 := r.Group(conf.App.RoutePrefix)
	store := persistence.NewInMemoryStore(BSCPerBlockTime)

	var (
		account     = controler.NewAccountController(pendingTxs)
		record      = controler.NewRecordController(pendingTxs)
		inscription = controler.NewInscriptionController(pendingTxs)
	)

	v1 := bnb48.Group("/v1")
	{
		v1.POST("/balance/list", cache.CachePageAtomic(store, BSCPerBlockTime, account.List))
		v1.POST("/record", cache.CachePageAtomic(store, BSCPerBlockTime, record.Get))
		v1.POST("/record/list", cache.CachePageAtomic(store, BSCPerBlockTime, record.List))
		v1.POST("/inscription/list", cache.CachePageAtomic(store, BSCPerBlockTime, inscription.List))
		v1.POST("/account/balance", cache.CachePageAtomic(store, BSCPerBlockTime, account.Balance))
	}

	return r
}
