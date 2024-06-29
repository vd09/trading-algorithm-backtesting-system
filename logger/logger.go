package logger

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/vd09/trading-algorithm-backtesting-system/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerInterface defines the methods for a logger
type LoggerInterface interface {
	InitLogger()
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
	SetLogLevel(level string)
}

// ZapLogger is the implementation of LoggerInterface using zap
type ZapLogger struct {
	log  *zap.Logger
	once sync.Once
}

// InitLogger initializes the logger with optional file logging
func (zl *ZapLogger) InitLogger() {
	zl.once.Do(func() {
		logConfig := config.AppConfig
		logLevel := zl.getLogLevel(logConfig.Logger.Level)
		writeSyncer := zl.getLogWriter(logConfig.Logger.SaveInFile)
		core := zapcore.NewCore(zl.getEncoder(), writeSyncer, logLevel)

		zl.log = zap.New(core, zap.AddCaller())
		zap.ReplaceGlobals(zl.log)
	})
}

// getEncoder returns a new JSON encoder with custom configuration
func (zl *ZapLogger) getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter returns a WriteSyncer for logging to file or stdout
func (zl *ZapLogger) getLogWriter(logToFile bool) zapcore.WriteSyncer {
	if logToFile {
		file, err := os.OpenFile("logger/logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		return zapcore.AddSync(file)
	}
	return zapcore.AddSync(os.Stdout)
}

// getLogLevel returns the zapcore.Level based on the string level
func (zl *ZapLogger) getLogLevel(level string) zapcore.Level {
	levelMap := map[string]zapcore.Level{
		"debug":  zapcore.DebugLevel,
		"info":   zapcore.InfoLevel,
		"warn":   zapcore.WarnLevel,
		"error":  zapcore.ErrorLevel,
		"dpanic": zapcore.DPanicLevel,
		"panic":  zapcore.PanicLevel,
		"fatal":  zapcore.FatalLevel,
	}

	if logLevel, ok := levelMap[strings.ToLower(level)]; ok {
		return logLevel
	}
	return zapcore.InfoLevel
}

// Debug logs a message at Debug level with context
func (zl *ZapLogger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	zl.log.With(zapFieldsFromContext(ctx)...).Debug(msg, fields...)
}

// Info logs a message at Info level with context
func (zl *ZapLogger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	zl.log.With(zapFieldsFromContext(ctx)...).Info(msg, fields...)
}

// Warn logs a message at Warn level with context
func (zl *ZapLogger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	zl.log.With(zapFieldsFromContext(ctx)...).Warn(msg, fields...)
}

// Error logs a message at Error level with context
func (zl *ZapLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	zl.log.With(zapFieldsFromContext(ctx)...).Error(msg, fields...)
}

// Fatal logs a message at Fatal level with context
func (zl *ZapLogger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	zl.log.With(zapFieldsFromContext(ctx)...).Fatal(msg, fields...)
}

// SetLogLevel dynamically sets the log level
func (zl *ZapLogger) SetLogLevel(level string) {
	zl.log = zl.log.WithOptions(zap.IncreaseLevel(zl.getLogLevel(level)))
}

// zapFieldsFromContext extracts zap fields from the context
func zapFieldsFromContext(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	fields := []zap.Field{}
	if loggerFields, ok := ctx.Value("loggerFields").(map[string]interface{}); ok {
		for k, v := range loggerFields {
			fields = append(fields, zap.Any(k, v))
		}
	}
	return fields
}

// Global instance of the logger implementing LoggerInterface
var loggerInstance LoggerInterface = &ZapLogger{}

// InitLogger initializes the global logger instance
func InitLogger() {
	loggerInstance.InitLogger()
}

// GetLogger returns the global logger instance
func GetLogger() LoggerInterface {
	return loggerInstance
}
