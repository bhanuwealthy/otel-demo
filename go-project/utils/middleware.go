package utils

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
// totalReqCounter metric.Int64Counter
// responseTime    metric.Float64Histogram
)

func MetricMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer func() {
			// meter := monitoring.MeterProvider.Meter("http.server")
			reqOpts := getReqAttributes(r)
			totalReqCounter.Add(r.Context(), 1, reqOpts,
				metric.WithAttributeSet(attribute.NewSet(attribute.Int("http.response.status_code", 200))))
			responseTime.Record(r.Context(), int64(time.Since(startTime).Milliseconds()), reqOpts)
		}()
		next.ServeHTTP(w, r)
	})
}
