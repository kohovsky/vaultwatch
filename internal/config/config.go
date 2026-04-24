package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level configuration for vaultwatch.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	RoleID  string `yaml:"role_id"`
	SecretID string `yaml:"secret_id"`
}

// AlertsConfig defines alert thresholds and notification channels.
type AlertsConfig struct {
	Thresholds []string `yaml:"thresholds"` // e.g. ["72h", "24h", "1h"]
	Slack      *SlackConfig `yaml:"slack,omitempty"`
	Email      *EmailConfig `yaml:"email,omitempty"`
}

// SlackConfig holds Slack webhook settings.
type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
	Channel    string `yaml:"channel"`
}

// EmailConfig holds SMTP settings for email alerts.
type EmailConfig struct {
	SMTPHost   string   `yaml:"smtp_host"`
	SMTPPort   int      `yaml:"smtp_port"`
	From       string   `yaml:"from"`
	Recipients []string `yaml:"recipients"`
}

// MonitorConfig controls polling behavior.
type MonitorConfig struct {
	Interval  string   `yaml:"interval"`  // e.g. "5m"
	Paths     []string `yaml:"paths"`
}

// ParsedThresholds returns alert thresholds as time.Duration values.
func (a *AlertsConfig) ParsedThresholds() ([]time.Duration, error) {
	var durations []time.Duration
	for _, t := range a.Thresholds {
		d, err := time.ParseDuration(t)
		if err != nil {
			return nil, fmt.Errorf("invalid threshold %q: %w", t, err)
		}
		durations = append(durations, d)
	}
	return durations, nil
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}
	if c.Vault.Token == "" && (c.Vault.RoleID == "" || c.Vault.SecretID == "") {
		return fmt.Errorf("either vault.token or vault.role_id + vault.secret_id must be set")
	}
	if len(c.Monitor.Paths) == 0 {
		return fmt.Errorf("monitor.paths must contain at least one path")
	}
	return nil
}
