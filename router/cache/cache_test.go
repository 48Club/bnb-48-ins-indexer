package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	cache "github.com/chenyahui/gin-cache"
	persistence "github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	BSCPerBlockTime = 20 * time.Second
)

func TestXxx(t *testing.T) {
	testInitGin()
	time.Sleep(1 * time.Second)

	api := "http://127.0.0.1:8089/v1/a"

	{
		// check post cache, will cache
		body := []byte(`{"a":1}`)
		var cacheData string
		for i := 0; i < 10; i++ {
			resp, err := http.Post(api, "application/json", bytes.NewReader(body))
			assert.NoError(t, err)
			respData := i2json(t, resp.Body)
			if i == 0 {
				cacheData = respData
			}
			assert.Equal(t, cacheData, respData)
			time.Sleep(10 * time.Millisecond)
		}
	}
	{
		// check post cache, not cache
		var before string
		for i := 0; i < 10; i++ {
			resp, err := http.Post(api, "application/json", bytes.NewReader([]byte(fmt.Sprintf(`{"a":%d}`, i))))
			assert.NoError(t, err)
			respData := i2json(t, resp.Body)
			assert.NotEqual(t, before, respData)
			before = respData
			time.Sleep(10 * time.Millisecond)
		}
	}
	{
		// check post urlencoded cache, will cache
		body := []byte(`tick_hash=0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2&page=0&page_size=20`)
		var cacheData string
		for i := 0; i < 10; i++ {
			resp, err := http.Post(api, "application/x-www-form-urlencoded", bytes.NewReader(body))
			assert.NoError(t, err)
			respData := i2json(t, resp.Body)
			if i == 0 {
				cacheData = respData
			}
			assert.Equal(t, cacheData, respData)
			time.Sleep(10 * time.Millisecond)
		}
	}

	{
		// check post urlencoded cache, not cache
		var before string
		for i := 0; i < 10; i++ {
			resp, err := http.Post(api, "application/x-www-form-urlencoded", bytes.NewReader([]byte(fmt.Sprintf(`tick_hash=0xd893ca77b3122cb6c480da7f8a12cb82e19542076f5895f21446258dc473a7c2&page=%d&page_size=20`, i))))
			assert.NoError(t, err)
			respData := i2json(t, resp.Body)
			assert.NotEqual(t, before, respData)
			before = respData
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func i2json(t *testing.T, body io.ReadCloser) string {
	var s interface{}
	assert.NoError(t, json.NewDecoder(body).Decode(&s))
	assert.NoError(t, body.Close())
	b, err := json.Marshal(s)
	assert.NoError(t, err)
	return string(b)
}

func testInitGin() {
	r := gin.Default()

	store := persistence.NewMemoryStore(BSCPerBlockTime)

	v1 := r.Group("/v1", cache.Cache(store, BSCPerBlockTime, cache.WithCacheStrategyByRequest(func(c *gin.Context) (bool, cache.Strategy) {
		b, k := customCacheKeyGenerator(c)
		return b, cache.Strategy{
			CacheKey:      k,
			CacheDuration: BSCPerBlockTime,
		}
	})))
	{
		v1.POST("/a", func(ctx *gin.Context) {
			data, _ := io.ReadAll(ctx.Request.Body)
			defer ctx.Request.Body.Close()
			_ = ctx.Request.ParseForm()
			ctx.JSON(200, gin.H{
				"message": time.Now().UnixMilli(),
				"body":    string(data),
				"url":     ctx.Request.PostForm.Encode(),
			})
		})
	}

	go func() {
		_ = r.Run(":8089")
	}()
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
			if c.Request.ParseForm() != nil {
				return false, key
			}
			_key += c.Request.PostForm.Encode()
		}
	}

	return true, key
}
