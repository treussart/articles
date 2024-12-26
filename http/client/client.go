package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"github.com/sony/gobreaker/v2"
	"github.com/treussart/articles/http/client/circuitbreaker"
	"github.com/treussart/articles/http/client/retryable"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	// We need to consume response bodies to maintain http connections, but
	// limit the size we consume to respReadLimit.
	defaultConcurrency      = 100
	defaultTimeout          = 4 * time.Second
	defaultRetryMax         = 3
	defaultRetryWaitMin     = 50 * time.Millisecond
	defaultRetryWaitMax     = 1 * time.Second
	defaultKeepAliveTimeout = 15 * time.Second
	defaultCBTimeout        = 60 * time.Second
	// Circuit breaker does not take retries into account
	defaultCBConsecutiveFailures = 2
	defaultCBMaxRequests         = 1
	defaultCBHTTPSatusCodeMax    = http.StatusInternalServerError
)

func getTransport(config customConfig) http.RoundTripper {
	tr := &http.Transport{
		TLSClientConfig:   nil,
		ForceAttemptHTTP2: false,
		// https://tleyden.github.io/blog/2016/11/21/tuning-the-go-http-client-library-for-load-testing/
		MaxIdleConns:          config.concurrency,
		MaxIdleConnsPerHost:   config.concurrency,
		MaxConnsPerHost:       0,
		DisableCompression:    false,
		DisableKeepAlives:     config.disableKeepAlive,
		IdleConnTimeout:       config.keepAliveTimeout,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if config.dialer != nil {
		tr.DialContext = config.dialer
	}
	if config.insecureSkipVerify {
		//nolint: gosec
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if config.proxyHost != "" {
		tr.Proxy = http.ProxyURL(&url.URL{Host: config.proxyHost})
	} else {
		tr.Proxy = nil
	}

	retryableTransport := &retryable.Transport{
		Tripper:      tr,
		RetryMax:     config.retryMax,
		RetryWaitMin: config.retryWaitMin,
		RetryWaitMax: config.retryWaitMax,
		Stats:        config.retryStats,
		ModuleName:   config.moduleName,
	}

	otelhttpTransport := otelhttp.NewTransport(retryableTransport,
		otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			return otelhttptrace.NewClientTrace(ctx)
		}),
		otelhttp.WithFilter(OperationalEndpointFilter),
	)

	if config.enableCircuitBreaker {
		if config.cbConsecutiveFailures == 0 {
			config.cbConsecutiveFailures = defaultCBConsecutiveFailures
		}
		cbConf := gobreaker.Settings{
			Name:        "HTTP Circuit Breaker",
			Timeout:     config.cbTimeout,
			MaxRequests: config.cbMaxRequests,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= config.cbConsecutiveFailures
			},
		}
		circuitBreakerTransport := &circuitbreaker.Transport{
			Tripper: otelhttpTransport,
			//nolint: bodyclose
			Breaker:       gobreaker.NewCircuitBreaker[*http.Response](cbConf),
			Stats:         config.circuitBreakerStats,
			ModuleName:    config.moduleName,
			StatusCodeMax: config.cbSatusCodeMax,
		}

		return circuitBreakerTransport
	}
	return otelhttpTransport
}

// OperationalEndpointFilter filters out requests to operational endpoints like "health", and "ready".
func OperationalEndpointFilter(r *http.Request) bool {
	if strings.Contains(r.URL.Path, "health") ||
		strings.Contains(r.URL.Path, "ready") {
		return false
	}
	return true
}

// Client creates a http.Client with configurable options for timeout, concurrency, retries, keep-alive, and circuit breaker.
func Client(options ...CustomOption) *http.Client {
	defaults := []CustomOption{
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(defaultRetryMax),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(defaultRetryWaitMax),
		WithKeepAliveTimeout(defaultKeepAliveTimeout),
		WithDisableKeepAlive(false),
		WithEnableCircuitBreaker(false),
		WithCBTimeout(defaultCBTimeout),
		WithCBMaxRequests(defaultCBMaxRequests),
		WithCBHTTPSatusCodeMax(defaultCBHTTPSatusCodeMax),
	}
	var config customConfig
	for _, opt := range append(defaults, options...) {
		opt(&config)
	}

	client := &http.Client{
		Transport: getTransport(config),
		Timeout:   config.timeout,
	}
	return client
}
