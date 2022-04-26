package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress          string `env:"RUN_ADDRESS" envDefault:"localhost:8000"`
	DatabaseURL         string `env:"DATABASE_URI"`
	ActualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func (c *Config) Parse() error {
	if err := env.Parse(c); err != nil {
		return err
	}

	flag.StringVar(&c.RunAddress, "a", c.RunAddress, "server address")
	flag.StringVar(&c.DatabaseURL, "d", c.DatabaseURL, "database url")
	flag.StringVar(&c.ActualSystemAddress, "r", c.ActualSystemAddress, "actual system address")
	flag.Parse()

	return nil
}
