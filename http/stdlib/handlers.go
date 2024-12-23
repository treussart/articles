package main

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func taskHandler(logger zerolog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Perform task operation
		logger.Info().Msg("Performing task handler started...")
		time.Sleep(taskDuration / 2)
		time.Sleep(taskDuration / 2)
		logger.Info().Msg("Performing task handler ended...")
		w.WriteHeader(http.StatusOK)
	}
}

func taskHandlerWithDetectCancel(logger zerolog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		done := make(chan error)
		go func() {
			// Perform task operation
			logger.Info().Msg("Performing task handler started...")
			time.Sleep(taskDuration / 2)
			if r.Context().Err() != nil {
				done <- r.Context().Err()
			}
			time.Sleep(taskDuration / 2)
			logger.Info().Msg("Performing task handler ended...")
			done <- nil
		}()
		select {
		case <-r.Context().Done():
			logger.Warn().Err(r.Context().Err()).Msg("handleRequestCtx error")
			http.Error(w, r.Context().Err().Error(), http.StatusRequestTimeout)
		case err := <-done:
			if err != nil {
				logger.Error().Err(err).Msg("handleRequestCtx error")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			logger.Info().Msg("handleRequestCtx processed")
			w.WriteHeader(http.StatusOK)
		}
	}
}
