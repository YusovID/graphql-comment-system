package logger

import (
	"context"

	"go.uber.org/zap"
)

const(
	Key = "logger"
	RequestID = "requestID"
)

var logger *zap.Logger

type Logger struct {
	l *zap.Logger
}

func New(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	ctx= context.WithValue(ctx, Key, &Logger{logger})
	return ctx, nil
}

func GetLoggerFromContext(ctx context.Context) *Logger {
	return ctx.Value(Key).(*Logger)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String("requestID", ctx.Value(RequestID).(string)))
	}
	l.l.Info(msg, fields...)
}