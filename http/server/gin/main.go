package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

const (
	ServiceName  = "http/gin"
	taskDuration = 12 * time.Second
)

type Config struct {
	ServicePort    string        `env:"SERVICE_PORT" envDefault:"9000"`
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" envDefault:"20s"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" envDefault:"20s"`
	HandlerTimeout time.Duration `env:"HANDLER_TIMEOUT" envDefault:"5s"`
}

func NewConfig() (*Config, error) {
	config := new(Config)
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}
	return config, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// The stop function releases resources associated with it, so code should call stop as soon as the operations running in this Context complete and signals no longer need to be diverted to the context.
	defer stop()
	zerolog.DurationFieldUnit = time.Second
	logger := zerolog.New(os.Stdout).With().Str("service", ServiceName).Timestamp().Logger()

	config, err := NewConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to setup config")
	}

	mux := gin.New()
	loggerMiddleware, err := NewLogger(logger, ServiceName)
	if err != nil {
		logger.Fatal().Err(err).Msg("Unable to setup logger middleware")
	}
	mux.Use(loggerMiddleware)
	mux.Use(NewTimeout(config.HandlerTimeout, http.StatusRequestTimeout, "timeout request exceeds "+config.HandlerTimeout.String()))

	mux.GET("/taskcancel", taskHandlerWithDetectCancel(logger))

	server := &http.Server{
		Addr: ":" + config.ServicePort,
		Handler: recovery(
			mux,
			logger,
		),
		ErrorLog:     log.New(logger, "", log.LstdFlags),
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

	<-ctx.Done()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = server.Shutdown(ctxTimeout); err != nil {
		logger.Fatal().Err(err).Msg("server.Shutdown error")
	}
}
