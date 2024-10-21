package logger

import (
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

func init() {
	// Custo Zap encoder configuration
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339), // ISO8601 UTC
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// set up the logger
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// build the logger
	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Unable to initiate the logger: %s", err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()
}

func Info(args ...interface{}) {
	sugar.Info(args)
}

func Infof(message string, args ...interface{}) {
	sugar.Infof(message, args)
}

func Warnf(message string, args ...interface{}) {
	sugar.Warnf(message, args)
}

func Error(args ...interface{}) {
	sugar.Error(args)
}

func Errorf(message string, args ...interface{}) {
	sugar.Errorf(message, args)
}

func Fatalf(message string, args ...interface{}) {
	sugar.Fatalf(message, args)
}
