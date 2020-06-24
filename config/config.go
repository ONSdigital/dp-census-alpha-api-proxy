package config

import (
	"errors"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-census-alpha-api-proxy
type Config struct {
	BindAddr                string `envconfig:"BIND_ADDR"`
	AuthToken               string `envconfig:"AUTH_TOKEN" json:"-"`
	FlexibleTableBuilderURL string `envconfig:"FTB_URL"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                ":8080",
		AuthToken:               "",
		FlexibleTableBuilderURL: "http://localhost:8491/v6",
	}

	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.AuthToken) == 0 {
		return nil, errors.New("auth token cannot be empty")
	}

	return cfg, nil
}

func (c *Config) GetAuthToken() string {
	if !strings.HasPrefix("Bearer ", c.AuthToken) {
		c.AuthToken = "Bearer " + c.AuthToken
	}
	return c.AuthToken
}
