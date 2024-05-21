package router

import (
	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/controler"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/types"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	cache "github.com/chenyahui/gin-cache"
	persistence "github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	BSCPerBlockTime = 3 * time.Second
)

func NewBotRoute(pendingTxs *types.GlobalVariable) *gin.Engine {
	r := gin.Default()
	r.Use(cors)

	conf := config.GetConfig()
	bnb48 := r.Group(conf.App.RoutePrefix)

	store := persistence.NewMemoryStore(BSCPerBlockTime)

	var (
		account     = controler.NewAccountController(pendingTxs)
		record      = controler.NewRecordController(pendingTxs)
		inscription = controler.NewInscriptionController(pendingTxs)
		wrap        = controler.NewWrapController()
	)

	v1 := bnb48.Group("/v1", cache.Cache(store, BSCPerBlockTime, cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, cache.Strategy) {
		b, k := customCacheKeyGenerator(c)
		return b, cache.Strategy{
			CacheKey:      k,
			CacheDuration: BSCPerBlockTime,
		}
	})))
	{
		v1.POST("/balance/list", account.List)
		v1.POST("/record", record.Get)
		v1.POST("/record/list", record.List)
		v1.POST("/inscription/list", inscription.List)
		v1.POST("/account/balance", account.Balance)
		v1.POST("/wrap_unwrap/list", wrap.List)
	}

	return r
}

func customCacheKeyGenerator(c *gin.Context) (t bool, key string) {
	_key := c.Request.URL.Path
	defer func() {
		key = uuid.NewMD5(uuid.NameSpaceURL, []byte(_key)).String()
	}()
	if c.Request.Method == http.MethodPost {
		ct := c.Request.Header.Get("Content-Type")
		if strings.Contains(ct, "application/json") {
			data, _ := io.ReadAll(c.Request.Body)
			defer c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(data))
			_key += string(data)
		} else if strings.Contains(ct, "application/x-www-form-urlencoded") {
			if err := c.Request.ParseForm(); err != nil {
				log.Sugar.Errorf("gin cache key parse form error: %s", err.Error())
				return false, key
			}
			_key += c.Request.PostForm.Encode()
		}
	}

	return true, key
}

func cors(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
	}
	c.Next()
}
