package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full vaultwatch configuration.
type Config struct {
	VaultAddress string   `yaml:"vault_address"`
	VaultToken   string   `yaml:"vault_token"`
	Paths        []string `yaml:"paths"`
	Interval     string   `yaml:"interval"`
	Thresholds   struct {
		Warning  string `yaml:"warning"`
		Critical string `yaml:"critical"`
	} `yaml:"thresholds"`
	Alerts struct {
		LogFile    string `yaml:"log_file"`
		WebhookURL string `yaml:"webhook_url"`
		SlackURL   string `yaml:"slack_url"`
	} `yaml:"alerts"`

	// Parsed durations — populated by Load.
	WarningThreshold  time.Duration
	CriticalThreshold time.Duration
	PollInterval      time.Duration
}

// Load reads and validates a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: invalid YAML: %w", err)
	}

	if cfg.VaultAddress == "" {
		return nil, errors.New("config: vault_address is required")
	}
	if len(cfg.Paths) == 0 {
		return nil, errors.New("config: at least one path is required")
	}

	warning := cfg.Thresholds.Warning
	if warning == "" {
		warning = "72h"
	}
	critical := cfg.Thresholds.Critical
	if critical == "" {
		critical = "24h"
	}
	interval := cfg.Interval
	if interval == "" {
		interval = "5m"
	}

	cfg.WarningThreshold, err = time.ParseDuration(warning)
	if err != nil {
		return nil, fmt.Errorf("config: invalid warning threshold %q: %w", warning, err)
	}
	cfg.CriticalThreshold, err = time.ParseDuration(critical)
	if err != nil {
		return nil, fmt.Errorf("config: invalid critical threshold %q: %w", critical, err)
	}
	cfg.PollInterval, err = time.ParseDuration(interval)
	if err != nil {
		return nil, fmt.Errorf("config: invalid interval %q: %w", interval, err)
	}

	return &cfg, nil
}
