package main

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func healthHandler(logger zerolog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info().Msg("Performing health handler")
		w.WriteHeader(http.StatusOK)
	}
}

func readyHandler(logger zerolog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		logger.Info().Msg("Performing ready handler")
		w.WriteHeader(http.StatusOK)
	}
}

func taskHandler(logger zerolog.Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		// Perform task operation
		logger.Info().Msg("Performing task handler started...")
		time.Sleep(taskDuration)
		logger.Info().Msg("Performing task handler ended...")
		w.WriteHeader(http.StatusOK)
	}
}
