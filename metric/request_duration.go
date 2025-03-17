package metric

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.20.0/httpconv"
)

const (
	metricNameRequestDurationMs = "request_duration_millis"
	metricUnitRequestDurationMs = "ms"
	metricDescRequestDurationMs = "Measures the latency of HTTP requests processed by the server, in milliseconds."
)

func NewRequestDurationMillis(cfg BaseConfig) func(next http.Handler) http.Handler {
	// init metric, here we are using histogram for capturing request duration
	histogram, err := cfg.Meter.Int64Histogram(
		metricNameRequestDurationMs,
		otelmetric.WithDescription(metricDescRequestDurationMs),
		otelmetric.WithUnit(metricUnitRequestDurationMs),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to create %s histogram: %v", metricNameRequestDurationMs, err))
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// capture the start time of the request
			startTime := time.Now()

			// execute next http handler
			next.ServeHTTP(w, r)

			// record the request duration
			duration := time.Since(startTime)

			// Get attributes from ServerRequest
			serverAttrs := httpconv.ServerRequest(cfg.ServerName, r)
			// Prepend http_target attribute
			attrs := append([]attribute.KeyValue{attribute.String("http_target", r.URL.RequestURI())}, serverAttrs...)

			histogram.Record(
				r.Context(),
				int64(duration.Milliseconds()),
				otelmetric.WithAttributes(attrs...),
			)
		})
	}
}
