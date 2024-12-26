package retryable

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/treussart/articles/http/client/metrics"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

const (
	defaultRespReadLimit = int64(4096)
)

var (
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	// A regular expression to match the error returned by net/http when a
	// request header or value is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	invalidHeaderErrorRe = regexp.MustCompile(`invalid header`)

	// A regular expression to match the error returned by net/http when the
	// TLS certificate is not trusted. This error isn't typed
	// specifically so we resort to matching on the error string.
	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

// Transport handles HTTP transactions with configurable retry policy and stats recording.
type Transport struct {
	Tripper      http.RoundTripper
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	Stats        *Stats
	ModuleName   string
}

// parseRetryAfterHeader parses the Retry-After header and returns the
// delay duration according to the spec: https://httpwg.org/specs/rfc7231.html#header.retry-after
// The bool returned will be true if the header was successfully parsed.
// Otherwise, the header was either not present, or was not parseable according to the spec.
//
// Retry-After headers come in two flavors: Seconds or HTTP-Date
//
// Examples:
// * Retry-After: Fri, 31 Dec 1999 23:59:59 GMT
// * Retry-After: 120
func parseRetryAfterHeader(headers []string) (time.Duration, bool) {
	if len(headers) == 0 || headers[0] == "" {
		return 0, false
	}
	header := headers[0]
	// Retry-After: 120
	if sleep, err := strconv.ParseInt(header, 10, 64); err == nil {
		if sleep < 0 { // a negative sleep doesn't make sense
			return 0, false
		}
		return time.Second * time.Duration(sleep), true
	}

	// Retry-After: Fri, 31 Dec 1999 23:59:59 GMT
	retryTime, err := time.Parse(time.RFC1123, header)
	if err != nil {
		return 0, false
	}
	if until := time.Until(retryTime); until > 0 {
		return until, true
	}
	// date is in the past
	return 0, true
}

// DefaultBackoff provides a default callback for Client.Backoff which
// will perform exponential backoff based on the attempt number and limited
// by the provided minimum and maximum durations.
func backoff(minimum, maximum time.Duration, retries int, resp *http.Response) time.Duration {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if sleep, ok := parseRetryAfterHeader(resp.Header["Retry-After"]); ok {
				return sleep
			}
		}
	}
	mult := math.Pow(2, float64(retries)) * float64(minimum)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > maximum {
		sleep = maximum
	}
	return sleep
}

func isCertError(err error) bool {
	var certificateVerificationError *tls.CertificateVerificationError
	ok := errors.As(err, &certificateVerificationError)
	return ok
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		var v *url.Error
		if errors.As(err, &v) {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false
			}

			// Don't retry if the error was due to an invalid header.
			if invalidHeaderErrorRe.MatchString(v.Error()) {
				return false
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if notTrustedErrorRe.MatchString(v.Error()) {
				return false
			}
			if isCertError(v.Err) {
				return false
			}
		}
		return true
	}

	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side. This will catch
	// invalid response codes as well, like 0 and 999.
	if resp.StatusCode == 0 || (resp.StatusCode >= http.StatusInternalServerError && resp.StatusCode != http.StatusNotImplemented) {
		return true
	}

	return false
}

func drainBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, defaultRespReadLimit))
		_ = resp.Body.Close()
	}
}

// RoundTrip executes a single HTTP transaction and retries on failure based on the configured retry policy.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	if t.Stats != nil {
		defer t.Stats.Duration.Record(context.Background(), metrics.MeasureDuration(start), api.WithAttributes(
			attribute.String(metrics.PKGLabelName, t.ModuleName),
		))
	}

	// Clone the request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Send the request
	resp, err := t.Tripper.RoundTrip(req)

	// Retry logic
	retries := 0
	for shouldRetry(err, resp) && retries < t.RetryMax {
		if t.Stats != nil {
			t.Stats.Retry.Add(context.Background(), 1, api.WithAttributes(
				attribute.String(metrics.PKGLabelName, t.ModuleName)))
		}
		// Wait for the specified backoff period
		time.Sleep(backoff(t.RetryWaitMin, t.RetryWaitMax, retries, resp))

		// We're going to retry, consume any response to reuse the connection.
		drainBody(resp)

		// Retry the request
		resp, err = t.Tripper.RoundTrip(req)

		retries++
	}
	if err != nil {
		return resp, fmt.Errorf("t.transport.RoundTrip: %w", err)
	}

	// Return the response
	return resp, nil
}
