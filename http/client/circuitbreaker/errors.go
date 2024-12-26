package circuitbreaker

import "errors"

var ErrHTTP = errors.New("circuit breaker unexpected HTTP error")
var ErrUnexpectedHTTPStatus = errors.New("unexpected HTTP status")
