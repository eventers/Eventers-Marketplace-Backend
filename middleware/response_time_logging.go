package middleware

import (
	"fmt"
	"net/http"
	"time"

	"eventers-marketplace-backend/logger"
)

func ResponseTimeLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer logger.LogExecutionTime(r.Context(), time.Now().UTC(), fmt.Sprintf("Total response for %s", r.URL.Path))
		next.ServeHTTP(w, r)
	})
}
