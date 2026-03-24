package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"expvar"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const requestIDHeader = "X-Request-ID"

const requestIDContextKey = contextKey("request_id")

var (
	totalRequestsMetric = expvar.NewInt("total_requests")
	responses2xxMetric  = expvar.NewInt("responses_2xx_total")
	responses4xxMetric  = expvar.NewInt("responses_4xx_total")
	responses5xxMetric  = expvar.NewInt("responses_5xx_total")
	dbStatsProviderMu   sync.RWMutex
	dbStatsProvider     = func() sql.DBStats { return sql.DBStats{} }

	httpRequestsTotalMetric = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relohelper_http_requests_total",
			Help: "Total number of HTTP requests processed by the API.",
		},
		[]string{"method", "route"},
	)
	httpRequestsByStatusClassMetric = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "relohelper_http_requests_by_status_class_total",
			Help: "Total number of HTTP requests partitioned by status class.",
		},
		[]string{"method", "route", "status_class"},
	)
	httpRequestDurationMetric = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "relohelper_http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status_class"},
	)
	rateLimiterRejectedMetric = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "relohelper_rate_limiter_rejected_total",
			Help: "Total number of HTTP requests rejected by the rate limiter.",
		},
	)
	rateLimiterAllowedMetric = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "relohelper_rate_limiter_allowed_total",
			Help: "Total number of HTTP requests allowed by the rate limiter.",
		},
	)
	dbOpenConnectionsMetric = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "relohelper_db_open_connections",
			Help: "Current number of open database connections.",
		},
		func() float64 {
			return float64(getDBStats().OpenConnections)
		},
	)
	dbInUseConnectionsMetric = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "relohelper_db_in_use_connections",
			Help: "Current number of database connections in use.",
		},
		func() float64 {
			return float64(getDBStats().InUse)
		},
	)
	dbIdleConnectionsMetric = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "relohelper_db_idle_connections",
			Help: "Current number of idle database connections.",
		},
		func() float64 {
			return float64(getDBStats().Idle)
		},
	)
	dbWaitCountMetric = promauto.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "relohelper_db_wait_count_total",
			Help: "Total number of waits for a database connection.",
		},
		func() float64 {
			return float64(getDBStats().WaitCount)
		},
	)
	dbWaitDurationMetric = promauto.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "relohelper_db_wait_duration_seconds_total",
			Help: "Total time blocked waiting for a database connection.",
		},
		func() float64 {
			return getDBStats().WaitDuration.Seconds()
		},
	)
	dbMaxOpenConnectionsMetric = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "relohelper_db_max_open_connections",
			Help: "Configured maximum number of open database connections.",
		},
		func() float64 {
			return float64(getDBStats().MaxOpenConnections)
		},
	)
)

var _ = []any{
	dbOpenConnectionsMetric,
	dbInUseConnectionsMetric,
	dbIdleConnectionsMetric,
	dbWaitCountMetric,
	dbWaitDurationMetric,
	dbMaxOpenConnectionsMetric,
}

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

func (app *application) collectPrometheusMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(recorder, r)

		if r.URL.Path == "/metrics" || r.URL.Path == "/favicon.ico" {
			return
		}

		route := requestRoute(r)
		statusClass := strconv.Itoa(recorder.statusCode/100) + "xx"
		labels := prometheus.Labels{
			"method": r.Method,
			"route":  route,
		}

		httpRequestsTotalMetric.With(labels).Inc()
		httpRequestsByStatusClassMetric.With(prometheus.Labels{
			"method":       r.Method,
			"route":        route,
			"status_class": statusClass,
		}).Inc()
		httpRequestDurationMetric.With(prometheus.Labels{
			"method":       r.Method,
			"route":        route,
			"status_class": statusClass,
		}).Observe(time.Since(start).Seconds())
	})
}

func requestRoute(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		return "unmatched"
	}

	routePattern := routeContext.RoutePattern()
	if routePattern != "" {
		return requestRouteMode(routePattern, r)
	}

	return "unmatched"
}

func requestRouteMode(routePattern string, r *http.Request) string {
	qs := r.URL.Query()

	switch routePattern {
	case "/cities":
		if qs.Has("ids") {
			if hasDetailedCityIncludeQuery(qs) {
				return "/cities:batch_detailed"
			}
			return "/cities:batch"
		}
		return "/cities:list"
	case "/cities/{id}":
		if qs.Has("include") {
			return "/cities/{id}:detailed"
		}
		return routePattern
	case "/countries":
		if qs.Has("country_codes") {
			if qs.Has("include") {
				return "/countries:batch_detailed"
			}
			return "/countries:batch"
		}
		return "/countries:list"
	case "/countries/{alpha3}":
		if qs.Has("include") {
			return "/countries/{alpha3}:detailed"
		}
		return routePattern
	default:
		return routePattern
	}
}

func hasDetailedCityIncludeQuery(qs url.Values) bool {
	rawInclude := qs.Get("include")
	if rawInclude == "" {
		return false
	}

	for _, part := range strings.Split(rawInclude, ",") {
		switch strings.TrimSpace(part) {
		case "numbeo_cost", "numbeo_indices", "avg_climate":
			return true
		}
	}

	return false
}

func newRequestID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}

	return hex.EncodeToString(b[:])
}

func setDBStatsProvider(db *sql.DB) {
	dbStatsProviderMu.Lock()
	defer dbStatsProviderMu.Unlock()

	dbStatsProvider = func() sql.DBStats {
		if db == nil {
			return sql.DBStats{}
		}

		return db.Stats()
	}
}

func getDBStats() sql.DBStats {
	dbStatsProviderMu.RLock()
	provider := dbStatsProvider
	dbStatsProviderMu.RUnlock()

	return provider()
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
