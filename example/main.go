package main

import (
	log "github.com/caser789/logger"
	"go.uber.org/zap"
	"time"
)

func main() {
	logger := log.GetLogger()
	sysLogger := log.GetSysLogger()

	//c := make(chan struct{})

	defer func() {
		logger.Sync()
		sysLogger.Sync()
	}()

	func() {
		for i := 0; i < 2; i++ {
			logger.Info("hello", zap.String("test", "test"))
			logger.Error("world", zap.Int("one", 1))
			sysLogger.Info("good", zap.String("test", "test"))
			sysLogger.Error("bye", zap.Int("one", 1))

			time.Sleep(1 * time.Second)
		}
		log.SetLevel(zap.ErrorLevel, 3*time.Second)
		for i := 0; i < 1; i++ {
			logger.Info("hello", zap.String("test", "test"))
			logger.Error("world", zap.Int("one", 1))
			sysLogger.Info("good", zap.String("test", "test"))
			sysLogger.Error("bye", zap.Int("one", 1))

			time.Sleep(1 * time.Second)
		}
		log.SetLevel(zap.DebugLevel, 0*time.Second)
		for i := 0; i < 1; i++ {
			logger.Info("hello", zap.String("test", "test"))
			logger.Error("world", zap.Int("one", 1))
			sysLogger.Info("good", zap.String("test", "test"))
			sysLogger.Error("bye", zap.Int("one", 1))

			time.Sleep(1 * time.Second)
		}
	}()
}
