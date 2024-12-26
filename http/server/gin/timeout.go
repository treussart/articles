package main

import (
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

var ErrCancelCauseTimeout = errors.New("cancel cause Gin timeout")

// NewTimeout creates a middleware that adds a timeout to each request, responding with a specified code and message on timeout.
func NewTimeout(timeout time.Duration, responseCodeTimeout int, responseBodyTimeout string) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeoutCause(c.Request.Context(), timeout, ErrCancelCauseTimeout)

		defer func() {
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				c.String(responseCodeTimeout, responseBodyTimeout)
				c.Abort()
			}

			cancel()
		}()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
