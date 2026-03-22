package http

import (
	"bytes"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func ZapLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			writer := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			defer func() {
				requestLogger := logger.With(
					zap.String("method", r.Method),
					zap.Int("status", writer.Status()),
					zap.String("path", r.URL.Path),
					zap.String("request_id", middleware.GetReqID(r.Context())),
					zap.Duration("latency", time.Since(start)),
					zap.Int("response_size", writer.BytesWritten()),
				)
				if writer.Status() >= http.StatusInternalServerError {
					var buffer bytes.Buffer
					writer.Tee(&buffer)

					requestLogger.Error(
						"error",
						zap.String("response", buffer.String()),
					)
				} else {
					requestLogger.Info("request")
				}
			}()

			next.ServeHTTP(writer, r)
		}

		return http.HandlerFunc(fn)
	}
}
