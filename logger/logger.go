package logger

import (
	"os"
	"strings"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.Logger

// InitLogger initializes the logger with optional file logging
func InitLogger() {
	logConfig := config.AppConfig
	logLevel := getLogLevel(logConfig.Logger.Level)
	writeSyncer := getLogWriter(logConfig.Logger.SaveInFile)
	core := zapcore.NewCore(getEncoder(), writeSyncer, logLevel)

	log = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(log)
}

// getEncoder returns a new JSON encoder with custom configuration
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter returns a WriteSyncer for logging to file or stdout
func getLogWriter(logToFile bool) zapcore.WriteSyncer {
	if logToFile {
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logger/logs/app.log",
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		})
	}
	return zapcore.AddSync(os.Stdout)
}

// getLogLevel returns the zapcore.Level based on the string level
func getLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return log
}
