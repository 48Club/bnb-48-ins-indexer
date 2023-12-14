package log

import (
	"fmt"
	"github.com/jwrookie/fans/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"io"
)

var (
	Log   *zap.Logger
	Sugar *zap.SugaredLogger
	Write io.Writer
)

func Init(filename string) {
	ws, level := getConfigLogArgs(filename)
	encoderConf := zap.NewProductionEncoderConfig()
	encoderConf.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConf.EncodeCaller = CallerEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConf)
	log := zap.New(
		zapcore.NewCore(encoder, ws, zap.NewAtomicLevelAt(level)),
		zap.AddCaller(),
	)
	Log = log
	Sugar = log.Sugar()
}

func getConfigLogArgs(filename string) (zapcore.WriteSyncer, zapcore.Level) {
	log := config.GetConfig().Log
	level := zap.InfoLevel

	switch log.Level {
	case "ERROR", "errs":
		level = zap.ErrorLevel
	case "WARN", "warn":
		level = zap.WarnLevel
	case "", "INFO", "info":
		level = zap.InfoLevel
	case "DEBUG", "debug":
		level = zap.DebugLevel
	}

	var syncers []zapcore.WriteSyncer

	if filename != "" {
		logger := &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s", log.Dir, filename), // if logs dir not exist, it will be auto create
			MaxSize:    log.MaxSize,
			MaxBackups: log.MaxBackups,
			MaxAge:     log.MaxAge,
			Compress:   log.Compress,
			LocalTime:  true,
		}
		Write = logger
		syncers = append(syncers, zapcore.AddSync(logger))
	}

	//syncers = append(syncers, os.Stdout)
	ws := zapcore.NewMultiWriteSyncer(syncers...)

	return ws, level
}
