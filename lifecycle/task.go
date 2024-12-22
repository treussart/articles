package main

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

func performTask(ctx context.Context, wg *sync.WaitGroup, logger zerolog.Logger) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Task cancelled")
			return
		default:
			// Perform task operation
			logger.Info().Msg("Performing task started...")
			time.Sleep(taskDuration)
			logger.Info().Msg("Performing task ended...")
		}
	}
}
