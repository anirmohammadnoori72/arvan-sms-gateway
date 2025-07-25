package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	once sync.Once
)

func InitLogger() {
	once.Do(func() {
		var err error
		log, err = zap.NewProduction() // یا zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
	})
}

func Info(msg string, fields ...zap.Field) {
	if log == nil {
		InitLogger()
	}
	log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if log == nil {
		InitLogger()
	}
	log.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if log == nil {
		InitLogger()
	}
	log.Warn(msg, fields...)
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
