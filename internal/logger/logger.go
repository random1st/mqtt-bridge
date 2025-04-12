package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

func InitLogger(levelStr string) {
	once.Do(func() {
		levelStr = strings.ToLower(levelStr)

		var zapLevel zapcore.Level
		switch levelStr {
		case "debug":
			zapLevel = zapcore.DebugLevel
		case "info":
			zapLevel = zapcore.InfoLevel
		case "warn":
			zapLevel = zapcore.WarnLevel
		case "error":
			zapLevel = zapcore.ErrorLevel
		default:
			zapLevel = zapcore.InfoLevel
		}

		var cfg zap.Config
		if zapLevel == zapcore.DebugLevel {
			cfg = zap.NewDevelopmentConfig()
		} else {
			cfg = zap.NewProductionConfig()
		}

		cfg.Level.SetLevel(zapLevel)

		var err error
		logger, err = cfg.Build()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cannot initialize zap logger: %v\n", err)
			os.Exit(1)
		}

		logger.Info("Logger initialized",
			zap.String("level", zapLevel.String()))
	})
}

func L() *zap.Logger {
	if logger == nil {
		tmpLogger := zap.NewNop()
		return tmpLogger
	}
	return logger
}

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
