package middleware

import (
	"eventers-marketplace-backend/logger"
	"net/http"
)

func RequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf(r.Context(), "Request - %s %s, Headers: %+v", r.Method, r.URL, r.Header)
		next.ServeHTTP(w, r)
	})
}
