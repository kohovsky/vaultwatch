package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all vaultwatch configuration.
type Config struct {
	VaultAddress string        `yaml:"vault_address"`
	VaultToken   string        `yaml:"vault_token"`
	Paths        []string      `yaml:"paths"`
	Thresholds   []string      `yaml:"thresholds"`
	Interval     string        `yaml:"interval"`
	HistoryDir   string        `yaml:"history_dir"`
	Alerts       AlertsConfig  `yaml:"alerts"`
	Filter       FilterConfig  `yaml:"filter"`
}

// FilterConfig mirrors monitor.FilterConfig for YAML unmarshalling.
type FilterConfig struct {
	IncludePrefixes []string `yaml:"include_prefixes"`
	ExcludePrefixes []string `yaml:"exclude_prefixes"`
}

// AlertsConfig holds optional alert destinations.
type AlertsConfig struct {
	File    string `yaml:"file"`
	Webhook string `yaml:"webhook"`
	Slack   string `yaml:"slack"`
	Email   *EmailConfig `yaml:"email"`
}

// EmailConfig holds SMTP settings.
type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load reads and validates a YAML config file from path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.VaultAddress == "" {
		return nil, errors.New("vault_address is required")
	}
	if len(cfg.Paths) == 0 {
		return nil, errors.New("at least one path is required")
	}

	return &cfg, nil
}

// ParsedThresholds converts threshold strings (e.g. "24h") to durations.
func (c *Config) ParsedThresholds() ([]time.Duration, error) {
	defaults := []time.Duration{72 * time.Hour, 24 * time.Hour}
	if len(c.Thresholds) == 0 {
		return defaults, nil
	}
	out := make([]time.Duration, 0, len(c.Thresholds))
	for _, s := range c.Thresholds {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("invalid threshold %q: %w", s, err)
		}
		out = append(out, d)
	}
	return out, nil
}
