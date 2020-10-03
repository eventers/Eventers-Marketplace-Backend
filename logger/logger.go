package logger

import (
	"context"
	c "eventers-marketplace-backend/context"
	"fmt"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	logrus.Logger
}

var logger *logrus.Logger

const CorrelationId = "correlation_id"

func init() {
	logger = logrus.New()
	logger.SetOutput(os.Stdout)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	logger.WithField(CorrelationId, c.GetContextValue(ctx, c.ContextKeyCorrelationID)).Fatalf(format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	logger.WithField(CorrelationId, c.GetContextValue(ctx, c.ContextKeyCorrelationID)).Infof(format, args...)
}

func Info(ctx context.Context, msg string) {
	logger.WithField(CorrelationId, c.GetContextValue(ctx, c.ContextKeyCorrelationID)).Info(msg)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	formattedError := escapeString(format, args...)
	logger.WithField(CorrelationId, c.GetContextValue(ctx, c.ContextKeyCorrelationID)).Debug(formattedError)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	logger.WithField(CorrelationId, c.GetContextValue(ctx, c.ContextKeyCorrelationID)).Warnf(format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	formattedError := escapeString(format, args...)
	logger.WithField(CorrelationId, c.GetContextValue(ctx, c.ContextKeyCorrelationID)).Error(formattedError)
}

func escapeString(format string, args ...interface{}) string {
	errorMessage := fmt.Sprintf(format, args...)
	re := regexp.MustCompile(`(\n)|(\r\n)`)
	formattedError := re.ReplaceAllString(errorMessage, "\\n ")
	return formattedError
}
