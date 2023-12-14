package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jwrookie/fans/config"
	"github.com/jwrookie/fans/pkg/database"
	"github.com/jwrookie/fans/pkg/log"
	"github.com/jwrookie/fans/router"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
