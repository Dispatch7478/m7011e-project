package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "teamservice",
		Subsystem: "http",
		Name:      "in_flight_requests",
		Help:      "Current number of in-flight HTTP requests.",
	})

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "teamservice",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "teamservice",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
		},
		[]string{"method", "route", "code"},
	)
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		httpInFlight.Inc()
		start := time.Now()
		next.ServeHTTP(rec, r)
		httpInFlight.Dec()

		route := "unknown"
		if cr := mux.CurrentRoute(r); cr != nil {
			if t, err := cr.GetPathTemplate(); err == nil {
				route = t
			}
		}

		code := strconv.Itoa(rec.status)
		httpRequestsTotal.WithLabelValues(r.Method, route, code).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route, code).Observe(time.Since(start).Seconds())
	})
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}
