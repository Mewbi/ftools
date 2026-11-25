package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
	"go.uber.org/fx"
)

var Module = fx.Module("config", fx.Provide(New))

type Config struct {
	fx.Out
	Server Server `yaml:"server"`
	PDF    PDF    `yaml:"pdf"`
}

type Server struct {
	Addr string `yaml:"addr"`
}

type PDF struct {
	Timeout       time.Duration `yaml:"timeout"`
	QueueTimeout  time.Duration `yaml:"queue_timeout"`
	MaxConcurrent int           `yaml:"max_concurrent"` // 0 = number of CPUs
}

func New() (Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config/config.yaml"
	}

	cfg, err := load(path)
	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
}

func load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while reading config file: %w", err)
	}

	cfg := &Config{}
	dec := yaml.NewDecoder(bytes.NewReader(data))

	if err := dec.Decode(cfg); err != nil {
		return nil, fmt.Errorf("error while parsing config file: %w", err)
	}

	return cfg, nil
}
