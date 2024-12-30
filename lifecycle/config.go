package main

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	ServicePort    string        `env:"SERVICE_PORT" envDefault:"9000" validate:"required"`
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" envDefault:"20s"  validate:"required"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" envDefault:"20s"`
	HandlerTimeout time.Duration `env:"HANDLER_TIMEOUT" envDefault:"12s"`
}

func NewConfig() (*Config, error) {
	config := new(Config)
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}

	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("validate.Struct: %w", err)
	}

	return config, nil
}
