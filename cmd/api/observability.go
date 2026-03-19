package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"expvar"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const requestIDHeader = "X-Request-ID"

const requestIDContextKey = contextKey("request_id")

var (
	totalRequestsMetric = expvar.NewInt("total_requests")
	responses2xxMetric  = expvar.NewInt("responses_2xx_total")
	responses4xxMetric  = expvar.NewInt("responses_4xx_total")
	responses5xxMetric  = expvar.NewInt("responses_5xx_total")
)

func (app *application) contextSetRequestID(r *http.Request, requestID string) *http.Request {
	ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
	return r.WithContext(ctx)
}

func (app *application) contextGetRequestID(r *http.Request) string {
	requestID, ok := r.Context().Value(requestIDContextKey).(string)
	if !ok || requestID == "" {
		return ""
	}

	return requestID
}

func (app *application) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, app.contextSetRequestID(r, requestID))
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		requestID := app.contextGetRequestID(r)

		totalRequestsMetric.Add(1)
		switch recorder.statusCode / 100 {
		case 2:
			responses2xxMetric.Add(1)
		case 4:
			responses4xxMetric.Add(1)
		case 5:
			responses5xxMetric.Add(1)
		}

		if r.URL.Path == "/favicon.ico" {
			return
		}

		app.logger.Info(
			"http request",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.statusCode),
			slog.Duration("duration", duration),
			slog.String("remote_addr", r.RemoteAddr),
		)
	})
}

func requestRoute(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		return r.URL.Path
	}

	routePattern := routeContext.RoutePattern()
	if routePattern != "" {
		return routePattern
	}

	return r.URL.Path
}

func newRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}

	return hex.EncodeToString(b[:])
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
