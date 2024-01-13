package api

import (
	"bnb-48-ins-indexer/config"
	_ "bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/pkg/types"
	"bnb-48-ins-indexer/router"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Start(pendingTxs *types.GlobalVariable) {
	app := config.GetConfig().App
	log.Init("api.log")

	gin.DefaultWriter = log.Write
	r := router.NewBotRoute(pendingTxs)
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Port),
		Handler: r,
	}

	go func() {
		WaitForSignal(func() {
			ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
			defer cancel()
			if err := httpSrv.Shutdown(ctx); err != nil {
				log.Log.Error("failed to shutdown http server",
					zap.Error(err),
				)
			}
		})
	}()

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Log.Fatal("failed to run status server",
					zap.Error(err),
				)
			}
		}
	}()
}

func WaitForSignal(callback func()) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	sig := <-sigCh
	log.Log.Info("signal arrived",
		zap.String("signal", sig.String()),
	)

	callback()
}
