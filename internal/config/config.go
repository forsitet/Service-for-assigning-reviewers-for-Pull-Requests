package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type Config struct {
	HTTP *HTTPConfig `yaml:"http"`
}

func (c Config) HTTPAddr() string {
	if c.HTTP == nil {
		return ":8080"
	}
	return c.HTTP.Addr
}

func load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return &Config{}, fmt.Errorf("unmarshal config yaml: %w", err)
	}
	return cfg, nil
}

func ParseConfig() (*Config, error) {
	configPath := flag.String("config", "", "Path to config file")

	flag.Parse()

	if *configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}

	cfg, err := load(*configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
