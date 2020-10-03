package middleware

import (
	c "eventers-marketplace-backend/context"
	"eventers-marketplace-backend/logger"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type ContextKey string

func SetCorrelationIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("Correlation-Id")
		ctx := c.SetContextWithValue(r.Context(), c.ContextKeyCorrelationID, correlationID)
		if len(correlationID) == 0 {
			logger.Infof(ctx, "No correlation id provided. Generating a new one")
			correlationID = generateCorrelationID()
			ctx = c.SetContextWithValue(ctx, c.ContextKeyCorrelationID, correlationID)
			r.Header.Set("Correlation-Id", correlationID)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateCorrelationID() string {
	now := time.Now().UTC()
	secs := now.Unix()
	return fmt.Sprintf("%d.%d", rand.Int31(), secs)
}
