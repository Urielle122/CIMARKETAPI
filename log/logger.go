package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

func Init(fields ...zap.Field) *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	logLevel := "info"

	level, _ := zapcore.ParseLevel(logLevel)
	config.Level = zap.NewAtomicLevelAt(level)

	// Add this line to include caller information
	config.EncoderConfig.CallerKey = "file"

	// Build the logger with caller skipping
	logger, _ := config.Build(zap.AddCallerSkip(1))
	logger.With(fields...)

	defer logger.Sync()
	sugar = logger.With(fields...).Sugar()

	return sugar
}

// Info is now exported (capitalized)
func Info(msg string, keysAndValues ...interface{}) {
	sugar.Infow(msg, keysAndValues...)
}

func InfoF(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// Error is now exported (capitalized)
func Error(msg string, keysAndValues ...interface{}) {
	sugar.Errorw(msg, keysAndValues...)
}

func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}