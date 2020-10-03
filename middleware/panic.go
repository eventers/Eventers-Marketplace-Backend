package middleware

import (
	"eventers-marketplace-backend/response"
	"net/http"
	"runtime"
)

func PanicHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				const size = 1 << 16
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]

				response.SomethingWrong().Send(r.Context(), w)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
