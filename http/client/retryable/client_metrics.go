package retryable

import (
	"fmt"

	"github.com/treussart/articles/http/client/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Stats contains accumulated stats.
type Stats struct {
	Duration metric.Float64Histogram
	Retry    metric.Float64Counter
}

func GetStats(name string) (*Stats, error) {
	meter := otel.GetMeterProvider().Meter(name)
	duration, err := meter.Float64Histogram(
		metrics.Namespace+"client_http_response_seconds",
		metric.WithDescription("The duration in seconds of the http call"),
		metric.WithExplicitBucketBoundaries(.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10),
		metric.WithUnit("seconds"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Histogram: %w", err)
	}

	retry, err := meter.Float64Counter(metrics.Namespace+"client_http_retry_total",
		metric.WithDescription("Total number of retry that occurred"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}

	return &Stats{
		Duration: duration,
		Retry:    retry,
	}, nil
}
