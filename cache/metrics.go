package cache

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "dcache",
			Subsystem: "http",
			Name:      "request_latency_seconds",
			Help:      "Latency distribution of HTTP requests.",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2, 15), // Buckets from 0.5ms up to ~16s
		},
		[]string{"handler"},
	)
)

func init() {
	prometheus.MustRegister(requestLatency)
}

func InstrumentHandler(name string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		duration := time.Since(start).Seconds()
		requestLatency.WithLabelValues(name).Observe(duration)
	})
}

func ExposeMetrics(mux *http.ServeMux) {
	mux.Handle("/metrics", promhttp.Handler())
}
