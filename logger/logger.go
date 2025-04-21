package logger

import (
	"log"
	"os"
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

	// Get the log level from env
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		// Set default
		logLevel = "INFO"
	}
	zapLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		log.Fatalf("Wrong log level set: %s", err)
	}

	// set up the logger
	config := zap.Config{
		Level:            zapLevel,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// build the logger
	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		log.Fatalf("Unable to initiate the logger: %s", err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()
	sugar.Info("Initiated logger. Log level: ", logLevel)
}

func Info(args ...interface{}) {
	sugar.Info(args)
}

func Infof(message string, args ...interface{}) {
	sugar.Infof(message, args)
}

func Warn(args ...interface{}) {
	sugar.Warn(args)
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

func Fatal(args ...interface{}) {
	sugar.Fatal(args)
}

func Fatalf(message string, args ...interface{}) {
	sugar.Fatalf(message, args)
}

func Debug(args ...interface{}) {
	sugar.Debug(args)
}

func Debugf(message string, args ...interface{}) {
	sugar.Debugf(message, args)
}
