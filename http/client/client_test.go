package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sony/gobreaker/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/treussart/articles/http/client/circuitbreaker"
	"github.com/treussart/articles/http/client/dialer"
	"github.com/treussart/articles/http/client/retryable"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestClient(t *testing.T) {
	// http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(1),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(defaultRetryWaitMax),
	)
	response, err := httpClient.Get(svr.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	response, err = httpClient.Get("http://fake")
	require.Error(t, err)
	assert.Nil(t, response)
}

func TestClient_insecure(t *testing.T) {
	// https server for doh service
	svr := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	httpClient := Client(
		WithInsecureSkipVerify(true),
	)
	response, err := httpClient.Get(svr.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Create stdout exporter to be able to retrieve
	// the collected spans.
	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}

func TestClient_otel(t *testing.T) {
	tp, err := initTracer()
	if err != nil {
		require.NoError(t, err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	// http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Traceparent"))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(1),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(defaultRetryWaitMax),
	)
	response, err := httpClient.Get(svr.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestClient_otel_circuit_breaker(t *testing.T) {
	tp, err := initTracer()
	if err != nil {
		require.NoError(t, err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	// http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("Traceparent"))
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	// httpClient := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(1),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(defaultRetryWaitMax),
		WithEnableCircuitBreaker(true),
	)
	response, err := httpClient.Get(svr.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestClient_retry(t *testing.T) {
	// http server
	counter := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter++
		if counter == 3 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte("test"))
		}
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(2),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(1*time.Second),
	)
	start := time.Now()
	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	require.NoError(t, err)
	assert.Equal(t, 3, counter)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.LessOrEqual(t, defaultRetryWaitMin*2, time.Since(start))
}

func TestClient_timeout(t *testing.T) {
	// http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(1*time.Second),
		WithRetryMax(0),
	)
	start := time.Now()
	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Empty(t, response)
	assert.LessOrEqual(t, defaultRetryWaitMin*2, time.Since(start))
}

func TestClient_timeout_retry(t *testing.T) {
	// http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(1*time.Second),
		WithRetryMax(3),
		WithEnableCircuitBreaker(true),
		WithCBConsecutiveFailures(1),
	)
	start := time.Now()
	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	require.ErrorIs(t, err, context.DeadlineExceeded)
	fmt.Println(err.Error())
	assert.Empty(t, response)
	assert.LessOrEqual(t, defaultRetryWaitMin*2, time.Since(start))
}

func TestClient_dialer(t *testing.T) {
	// http server
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithDialer(dialer.GetDialer("8.8.8.8:53", 2*time.Second)),
	)
	response, err := httpClient.Get(svr.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)

	response, err = httpClient.Get("https://index.category.trustlane-qa.xyz/")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}

func TestClient_CB(t *testing.T) {
	// http server
	counter := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("test"))
	}))
	defer svr.Close()

	retryableStats, err := retryable.GetStats("ServiceName")
	require.NoError(t, err)
	circuitbreakerStats, err := circuitbreaker.GetStats("ServiceName")
	require.NoError(t, err)

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(0),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(1*time.Second),
		WithRetryableStats(retryableStats, "test"),
		WithCircuitBreakerStats(circuitbreakerStats, "test"),
		WithEnableCircuitBreaker(true),
		WithCBConsecutiveFailures(1),
		WithCBTimeout(2*time.Second),
	)
	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 1, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 1, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, gobreaker.ErrOpenState)

	time.Sleep(2 * time.Second)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 2, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)
}

func TestClient_CB_2(t *testing.T) {
	// http server
	counter := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("test"))
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(0),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(1*time.Second),
		WithEnableCircuitBreaker(true),
		WithCBConsecutiveFailures(2),
		WithCBTimeout(2*time.Second),
	)
	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 1, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 2, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 2, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, gobreaker.ErrOpenState)

	time.Sleep(2 * time.Second)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 3, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)
}

func TestClient_CB_retry(t *testing.T) {
	// http server
	counter := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("test"))
	}))
	defer svr.Close()

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(1),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(1*time.Second),
		WithEnableCircuitBreaker(true),
		WithCBConsecutiveFailures(1),
		WithCBTimeout(2*time.Second),
	)
	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 2, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 2, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, gobreaker.ErrOpenState)

	time.Sleep(2 * time.Second)

	response, err = httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 4, counter)
	assert.Nil(t, response)
	require.ErrorIs(t, err, circuitbreaker.ErrHTTP)
	require.ErrorIs(t, err, circuitbreaker.ErrUnexpectedHTTPStatus)
}

func TestClient_CB_target_error(t *testing.T) {
	// http server
	counter := 0
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		counter++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("test"))
	}))

	retryableStats, err := retryable.GetStats("ServiceName")
	require.NoError(t, err)
	circuitbreakerStats, err := circuitbreaker.GetStats("ServiceName")
	require.NoError(t, err)

	httpClient := Client(
		WithConcurrency(defaultConcurrency),
		WithTimeout(defaultTimeout),
		WithRetryMax(0),
		WithRetryWaitMin(defaultRetryWaitMin),
		WithRetryWaitMax(1*time.Second),
		WithRetryableStats(retryableStats, "test"),
		WithCircuitBreakerStats(circuitbreakerStats, "test"),
		WithEnableCircuitBreaker(true),
		WithCBConsecutiveFailures(1),
		WithCBTimeout(2*time.Second),
	)

	svr.Close()

	response, err := httpClient.Post(svr.URL, "text/html", strings.NewReader("test"))
	assert.Equal(t, 0, counter)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, circuitbreaker.ErrHTTP)
}
