package context

import (
	"context"
	"time"
)

const (
	ContextKeyCorrelationID ContextKey = "Correlation-Id"
	DefaultHttpTimeout                 = 30 * time.Second
)

type ContextKey string

func NewContextWithTimeOut(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

func NewContext(correlationID string) context.Context {
	return context.WithValue(context.Background(), ContextKeyCorrelationID, correlationID)
}

func SetContextWithValue(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetContextValue(ctx context.Context, key ContextKey) string {
	reqID := ctx.Value(key)
	if reqID != nil {
		if ret, ok := reqID.(string); ok {
			return ret
		}
	}
	return ""
}
