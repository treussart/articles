package circuitbreaker

import (
	"fmt"

	"github.com/treussart/articles/http/client/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Stats contains accumulated stats.
type Stats struct {
	CBOpen            metric.Float64Counter
	CBTooManyRequests metric.Float64Counter
}

func GetStats(name string) (*Stats, error) {
	meter := otel.GetMeterProvider().Meter(name)
	cbOpen, err := meter.Float64Counter(metrics.Namespace+"client_http_cbopen_total",
		metric.WithDescription("Total number of circuit-breaker openings"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}

	cbTooManyRequests, err := meter.Float64Counter(metrics.Namespace+"client_http_cbtoomanyrequests_total",
		metric.WithDescription("Total number of circuit-breaker openings"),
	)
	if err != nil {
		return nil, fmt.Errorf("meter.Float64Counter: %w", err)
	}

	return &Stats{
		CBOpen:            cbOpen,
		CBTooManyRequests: cbTooManyRequests,
	}, nil
}
