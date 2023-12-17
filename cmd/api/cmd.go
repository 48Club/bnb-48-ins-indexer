package api

import (
	"bnb-48-ins-indexer/config"
	"bnb-48-ins-indexer/pkg/database"
	"bnb-48-ins-indexer/pkg/log"
	"bnb-48-ins-indexer/router"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "api",
		Run: func(cmd *cobra.Command, args []string) {
			setup()
		},
	}
}

func setup() {
	app := config.GetConfig().App
	log.Init("api.log")
	database.NewMysql()

	gin.DefaultWriter = log.Write
	r := router.NewBotRoute()
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

	if err := httpSrv.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Log.Fatal("failed to run status server",
				zap.Error(err),
			)
		}
	}
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
