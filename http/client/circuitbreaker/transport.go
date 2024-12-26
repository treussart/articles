package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/sony/gobreaker/v2"
	"github.com/treussart/articles/http/client/metrics"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

// Transport represents an HTTP transport with integrated circuit-breaker and statistics tracking mechanisms.
type Transport struct {
	Tripper       http.RoundTripper
	Breaker       *gobreaker.CircuitBreaker[*http.Response]
	Stats         *Stats
	ModuleName    string
	StatusCodeMax int
}

// RoundTrip executes the HTTP request and returns the response or an error if the circuit breaker or the request fails.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	res, err := t.Breaker.Execute(func() (*http.Response, error) {
		res, err := t.Tripper.RoundTrip(r)
		if err != nil {
			return nil, fmt.Errorf("t.tripper.RoundTrip: %w", err)
		}

		if res != nil && res.StatusCode >= t.StatusCodeMax {
			return res, fmt.Errorf("status code %v: %w", res.StatusCode, ErrUnexpectedHTTPStatus)
		}

		return res, nil
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			if t.Stats != nil {
				t.Stats.CBOpen.Add(context.Background(), 1, api.WithAttributes(
					attribute.String(metrics.PKGLabelName, t.ModuleName),
				))
			}
		}
		if errors.Is(err, gobreaker.ErrTooManyRequests) {
			if t.Stats != nil {
				t.Stats.CBTooManyRequests.Add(context.Background(), 1, api.WithAttributes(
					attribute.String(metrics.PKGLabelName, t.ModuleName),
				))
			}
		}
		return nil, fmt.Errorf("t.breaker.Execute: %w: %w", ErrHTTP, err)
	}

	return res, nil
}
