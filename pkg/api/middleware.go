package api

import (
	"LoadBalancer/internal/logging"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		logging.L().Info("Request processed", zap.String("method", r.Method), zap.String("path", r.URL.Path), zap.Duration("duration", time.Since(start)), zap.String("remote_addr", r.RemoteAddr))
	})	
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		defer func() {
			if res := recover(); res != nil {
				logging.L().Error("Panic recovered", zap.Any("error", res))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}