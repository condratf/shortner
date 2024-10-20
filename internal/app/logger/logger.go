package logger

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// обёртка для ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.statusCode = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.contentLength += size
	return size, err
}

func InitLogger() *zap.SugaredLogger {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("Не удалось инициализировать логгер: %v", err)
	}
	return logger.Sugar()
}

func LoggingMiddleware(sugar *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			uri := r.RequestURI
			method := r.Method

			lw := &loggingResponseWriter{w, http.StatusOK, 0}

			next.ServeHTTP(lw, r)

			duration := time.Since(start)

			sugar.Infow("HTTP request",
				"method", method,
				"uri", uri,
				"status", lw.statusCode,
				"content_length", lw.contentLength,
				"duration", duration,
			)
		})
	}
}
