package middleware

import "net/http"

func SetContentTypeHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.Header().Add("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(responseWriter, request)
	})
}
