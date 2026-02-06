package config

import "errors"

type Config struct{}

func Load() (*Config, error) {
	return &Config{}, errors.New("not implemented")
}
