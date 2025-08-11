package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var appLogger *zap.Logger

func Init(environment string) error {
	if appLogger != nil {
		return nil
	}

	var config zap.Config
	if environment == "development" || environment == "dev" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	l, err := config.Build()
	if err != nil {
		return err
	}
	appLogger = l

	hostname, _ := os.Hostname()
	appLogger = appLogger.With(zap.String("host", hostname))

	return nil
}

func L() *zap.Logger {
	if appLogger == nil {
		return zap.NewNop()
	}
	return appLogger
}

func Sync() {
	if appLogger != nil {
		_ = appLogger.Sync()
	}
}
