package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog"
)

const (
	ServiceName  = "lifecycle"
	taskDuration = 10 * time.Second
)

type Config struct {
	ServicePort  string        `env:"SERVICE_PORT" envDefault:"9000"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"20s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"20s"`
}

func NewConfig() (*Config, error) {
	config := new(Config)
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}
	return config, nil
}

func main() {
	// Step 1: start
	startService := time.Now()
	// Create context that listens for the interrupt signal from the OS.
	// SIGINT (CRTL+C) only for testing purposes
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// The stop function releases resources associated with it, so code should call stop as soon as the operations running in this Context complete and signals no longer need to be diverted to the context.
	defer stop()
	zerolog.DurationFieldUnit = time.Second
	logger := zerolog.New(os.Stdout).With().Str("service", ServiceName).Timestamp().Logger()

	config, err := NewConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to setup config")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler(logger))
	mux.HandleFunc("/ready", readyHandler(logger))
	mux.HandleFunc("/task", taskHandler(logger))
	server := &http.Server{
		Addr:         ":" + config.ServicePort,
		Handler:      recovery(mux, logger),
		ErrorLog:     log.New(logger, "", 0),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}
	// Run HTTP server
	go func() {
		errListen := server.ListenAndServe()
		if err != nil && !errors.Is(errListen, http.ErrServerClosed) {
			logger.Fatal().Err(errListen).Msg("server.ListenAndServe error")
		}
	}()

	// WaitGroup waits for a collection of goroutines to finish, pass this by address
	waitGroup := &sync.WaitGroup{}

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		<-ctx.Done()
		logger.Info().Msg("HTTP server cancelled")
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// signal the HTTP server to stop
		errShutdown := server.Shutdown(timeoutCtx)
		if errShutdown != nil {
			logger.Error().Err(errShutdown).Msg("server.Shutdown error")
		}
	}()

	waitGroup.Add(1)
	go performTask(ctx, waitGroup, logger)
	logger.Info().Dur("duration", time.Since(startService)).Msg("Service started successfully")
	runningService := time.Now()
	// Step 2: running
	// Wait for termination signal
	<-ctx.Done()
	stop()
	// Step 3: shutdown
	// Start the graceful shutdown process
	logger.Info().Dur("duration", time.Since(runningService)).Msg("Gracefully shutting down service...")
	startGracefullyShuttingDown := time.Now()
	// wait for all goroutine finish their tasks.
	// it blocks until the WaitGroup counter is zero
	waitGroup.Wait()
	logger.Info().Dur("duration", time.Since(startGracefullyShuttingDown)).Msg("Shutdown service complete")
}
