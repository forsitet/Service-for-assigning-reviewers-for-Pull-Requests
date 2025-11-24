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

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type Config struct {
	HTTP *HTTPConfig    `yaml:"http"`
	DB   DatabaseConfig `yaml:"database"`
}

func (c Config) HTTPAddr() string {
	if c.HTTP == nil {
		return ":8080"
	}
	return c.HTTP.Addr
}

func (db DatabaseConfig) ConnString() string {
	host := db.Host
	if host == "" {
		host = "localhost"
	}

	port := db.Port
	if port == 0 {
		port = 5432
	}

	sslMode := db.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.User,
		db.Password,
		host,
		port,
		db.Name,
		sslMode,
	)
}

func load(path string) (*Config, error) {
	// #nosec G304 -- config file path is provided via command line flag
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return &Config{}, fmt.Errorf("unmarshal config yaml: %w", err)
	}

	if cfg.DB.User == "" || cfg.DB.Password == "" {
		return &Config{}, fmt.Errorf("database user and password must be set in config")
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
