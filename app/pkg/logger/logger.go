package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Key       = "logger"
	RequestID = "requestID"
)

type Logger struct {
	logger *zap.Logger
}

// New creates a new logger with the specified level and options.
func New(ctx context.Context, level zapcore.Level, opts ...zap.Option) (context.Context, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout", "logs/app.log"}        // Output to stdout and a file
	config.ErrorOutputPaths = []string{"stderr", "logs/error.log"} // Output errors to stderr and a file

	logger, err := config.Build(opts...)
	if err != nil {
		return ctx, fmt.Errorf("failed to create logger: %w", err)
	}

	//Ensure directory exists
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return ctx, fmt.Errorf("failed to create logs directory: %w", err)
	}
	defer logger.Sync()
	ctx = context.WithValue(ctx, Key, &Logger{logger})
	return ctx, nil
}

// GetLoggerFromContext retrieves a logger from the context.
func GetLoggerFromContext(ctx context.Context) (*Logger, error) {
	loggerInterface, ok := ctx.Value(Key).(*Logger)
	if !ok {
		return nil, fmt.Errorf("logger not found in context")
	}
	return loggerInterface, nil
}

// WithField adds a field to the logger's context.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{logger: l.logger.With(zap.Any(key, value))}
}

// Info logs an info message with the given fields.
func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.InfoLevel, msg, fields)
}

// Warn logs a warning message with the given fields.
func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.WarnLevel, msg, fields)
}

// Error logs an error message with the given fields.
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.ErrorLevel, msg, fields)
}

// Debug logs a debug message with the given fields.
func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.DebugLevel, msg, fields)
}

// Panic logs a panic message with the given fields.
func (l *Logger) Panic(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.PanicLevel, msg, fields)
}

// Fatal logs a fatal message with the given fields and then exits.
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.FatalLevel, msg, fields)
}

// log logs a message with the given level and fields.
func (l *Logger) log(ctx context.Context, level zapcore.Level, msg string, fields []zap.Field) {
	fields = append(fields, zap.String("time", time.Now().Format(time.RFC3339)))
	if ctx != nil && ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.logger.Log(level, msg, fields...)
}
