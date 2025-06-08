package middleware

import (
    "net/http"
    "time"

    "github.com/sirupsen/logrus"
)

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *logrus.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            next.ServeHTTP(rw, r)
            duration := time.Since(start)
            logger.WithFields(logrus.Fields{
                "method":     r.Method,
                "path":       r.URL.Path,
                "ip":         r.RemoteAddr,
                "status":     rw.statusCode,
                "duration_ms": duration.Milliseconds(),
            }).Info("Handled request")
        })
    }
}