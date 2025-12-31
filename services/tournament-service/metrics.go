package main

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "tournament_service",
		Subsystem: "http",
		Name:      "in_flight_requests",
		Help:      "Current number of in-flight HTTP requests.",
	})

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "tournament_service",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "tournament_service",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
		},
		[]string{"method", "route", "code"},
	)
)

// MetricsMiddleware for Echo
func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Skip tracking the /metrics endpoint itself
		if c.Path() == "/metrics" {
			return next(c)
		}

		httpInFlight.Inc()
		start := time.Now()

		err := next(c)

		httpInFlight.Dec()

		// Echo saves the matched route path (e.g., "/tournaments/:id") in c.Path()
		route := c.Path()
		if route == "" {
			route = "unknown"
		}

		status := c.Response().Status
		code := strconv.Itoa(status)

		// Handle errors that might not have set the status yet
		if err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				code = strconv.Itoa(httpErr.Code)
			} else {
				// 500 for generic errors
				if status == 200 { // Echo defaults to 200 even on error return sometimes
					code = "500"
				}
			}
		}

		httpRequestsTotal.WithLabelValues(c.Request().Method, route, code).Inc()
		httpRequestDuration.WithLabelValues(c.Request().Method, route, code).Observe(time.Since(start).Seconds())

		return err
	}
}

// Handler for the /metrics route
func MetricsHandler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}