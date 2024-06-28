package main

import (
	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"github.com/vd09/trading-algorithm-backtesting-system/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	config.InitConfig()
	logger.InitLogger()

	log := logger.GetLogger()
	log.Info("Starting the application...", zap.String("version", "1.0.0"))
	SomeFunction(log)
	// Rest of your application code
}

func SomeFunction(log *zap.Logger) {
	log.Debug("This is a debug message", zap.String("function", "SomeFunction"))
	log.Info("This is an info message", zap.Int("attempt", 3))
	log.Warn("This is a warning message", zap.String("warning", "potential issue"))
	//log.Error("This is an error message", zap.Error(err))
}
