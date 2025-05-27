package middleware

import (
	"net/http"
	"time"
	
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// RequestLogger returns a middleware that logs HTTP requests
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// Log request start
		log.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("request_id", middleware.GetReqID(r.Context())).
			Msg("Request started")
		
		// Process request
		next.ServeHTTP(ww, r)
		
		// Log request completion
		duration := time.Since(start)
		logger := log.Info()
		
		// Use error level for 5xx responses
		if ww.Status() >= 500 {
			logger = log.Error()
		} else if ww.Status() >= 400 {
			logger = log.Warn()
		}
		
		logger.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("request_id", middleware.GetReqID(r.Context())).
			Int("status", ww.Status()).
			Int("bytes_written", ww.BytesWritten()).
			Dur("duration", duration).
			Msg("Request completed")
	})
}

// Recovery returns a middleware that recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error().
					Interface("panic", rec).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("remote_addr", r.RemoteAddr).
					Str("request_id", middleware.GetReqID(r.Context())).
					Stack().
					Msg("Panic recovered")
				
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}