package controllers

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// statusRecorder captures the status code a handler writes, for logging
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// Unwrap lets http.ResponseController reach the underlying writer
func (s *statusRecorder) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

// LogRequests logs every request with its status and latency: errors for 5xx, warnings for 4xx
func LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		}
		slog.Log(r.Context(), level, "Request",
			"method", r.Method,
			"path", r.URL.RequestURI(),
			"status", rec.status,
			"latency", time.Since(start),
			"remote", r.RemoteAddr,
		)
	})
}

// Recover turns a panicking handler into a logged 500 instead of a dropped connection
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			// net/http uses this panic to abort a response on purpose
			if err == http.ErrAbortHandler {
				panic(err)
			}
			slog.Error("Panic serving request", "err", err, "path", r.URL.Path, "stack", string(debug.Stack()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}
