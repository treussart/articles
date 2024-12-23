package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func taskHandlerWithDetectCancel(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		contextLogger := zerolog.Ctx(c.Request.Context())
		done := make(chan error)
		go func() {
			// Perform task operation
			logger.Info().Msg("Performing task handler started...")
			time.Sleep(taskDuration / 2)
			if c.Request.Context().Err() != nil {
				done <- c.Request.Context().Err()
			}
			time.Sleep(taskDuration / 2)
			logger.Info().Msg("Performing task handler ended...")
			done <- nil
		}()
		select {
		case <-c.Request.Context().Done():
			contextLogger.UpdateContext(func(z zerolog.Context) zerolog.Context {
				return z.Err(c.Request.Context().Err())
			})
			c.Status(http.StatusRequestTimeout)
		case err := <-done:
			if err != nil {
				contextLogger.UpdateContext(func(z zerolog.Context) zerolog.Context {
					return z.Err(err)
				})
				c.Status(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusOK)
		}
	}
}
