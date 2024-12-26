package client

import (
	"context"
	"net"
	"time"

	"github.com/treussart/articles/http/client/circuitbreaker"
	"github.com/treussart/articles/http/client/retryable"
)

type customConfig struct {
	concurrency           int
	timeout               time.Duration
	retryMax              int
	retryWaitMin          time.Duration
	retryWaitMax          time.Duration
	retryStats            *retryable.Stats
	circuitBreakerStats   *circuitbreaker.Stats
	moduleName            string
	keepAliveTimeout      time.Duration
	disableKeepAlive      bool
	dialer                func(ctx context.Context, network string, addr string) (net.Conn, error)
	cbConsecutiveFailures uint32
	cbMaxRequests         uint32
	cbTimeout             time.Duration
	cbSatusCodeMax        int
	enableCircuitBreaker  bool
	insecureSkipVerify    bool
	proxyHost             string
}

type CustomOption func(*customConfig)

// WithConcurrency set max conns and idles per host.
func WithConcurrency(c int) CustomOption {
	return func(config *customConfig) {
		config.concurrency = c
	}
}

// WithTimeout set time duration for timeout.
func WithTimeout(d time.Duration) CustomOption {
	return func(config *customConfig) {
		config.timeout = d
	}
}

// WithRetryMax set max retry.
func WithRetryMax(c int) CustomOption {
	return func(config *customConfig) {
		config.retryMax = c
	}
}

// WithRetryWaitMin set time duration for minimum duration for retry.
func WithRetryWaitMin(d time.Duration) CustomOption {
	return func(config *customConfig) {
		config.retryWaitMin = d
	}
}

// WithRetryWaitMax set time duration for maximum duration for retry.
func WithRetryWaitMax(d time.Duration) CustomOption {
	return func(config *customConfig) {
		config.retryWaitMax = d
	}
}

// WithRetryableStats set stats and module name for metrics OTEL.
func WithRetryableStats(stats *retryable.Stats, moduleName string) CustomOption {
	return func(config *customConfig) {
		config.retryStats = stats
		config.moduleName = moduleName
	}
}

// WithCircuitBreakerStats set stats and module name for metrics OTEL.
func WithCircuitBreakerStats(stats *circuitbreaker.Stats, moduleName string) CustomOption {
	return func(config *customConfig) {
		config.circuitBreakerStats = stats
		config.moduleName = moduleName
	}
}

// WithKeepAliveTimeout set time duration for keep alive timeout.
func WithKeepAliveTimeout(d time.Duration) CustomOption {
	return func(config *customConfig) {
		config.keepAliveTimeout = d
	}
}

// WithDisableKeepAlive set true to disable keep-alive.
func WithDisableKeepAlive(d bool) CustomOption {
	return func(config *customConfig) {
		config.disableKeepAlive = d
	}
}

// WithDialer set dialer.
func WithDialer(dialer func(ctx context.Context, network string, addr string) (net.Conn, error)) CustomOption {
	return func(config *customConfig) {
		config.dialer = dialer
	}
}

// WithEnableCircuitBreaker for activate circuit breaker.
func WithEnableCircuitBreaker(enable bool) CustomOption {
	return func(config *customConfig) {
		config.enableCircuitBreaker = enable
	}
}

// WithCBConsecutiveFailures set circuit breaker consecutive failures/errors.
func WithCBConsecutiveFailures(failures uint32) CustomOption {
	return func(config *customConfig) {
		config.cbConsecutiveFailures = failures
	}
}

// WithCBTimeout set circuit breaker timeout, is the period of the open state, after which the state of the CircuitBreaker becomes half-open. If WithTimeout is less than or equal to 0, the timeout value of the CircuitBreaker is set to 60 seconds.
func WithCBTimeout(d time.Duration) CustomOption {
	return func(config *customConfig) {
		config.cbTimeout = d
	}
}

// WithCBMaxRequests is the maximum number of requests allowed to pass through when the CircuitBreaker is half-open. If MaxRequests is 0, the CircuitBreaker allows only 1 request.
func WithCBMaxRequests(d uint32) CustomOption {
	return func(config *customConfig) {
		config.cbMaxRequests = d
	}
}

// WithCBHTTPSatusCodeMax is the HTTP status code from which the CircuitBreaker counts errors.
func WithCBHTTPSatusCodeMax(d int) CustomOption {
	return func(config *customConfig) {
		config.cbSatusCodeMax = d
	}
}

// WithInsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
func WithInsecureSkipVerify(d bool) CustomOption {
	return func(config *customConfig) {
		config.insecureSkipVerify = d
	}
}

func WithProxyHost(d string) CustomOption {
	return func(config *customConfig) {
		config.proxyHost = d
	}
}
