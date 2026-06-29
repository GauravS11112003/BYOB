// Package config loads BYOB's declarative configuration: the data-source
// connectors to run and the AI provider settings.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration loaded from config.yaml.
type Config struct {
	Server     ServerConfig      `yaml:"server"`
	Connectors []ConnectorConfig `yaml:"connectors"`
	AI         AIConfig          `yaml:"ai"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Addr string `yaml:"addr"`
}

// ConnectorConfig is a single data-source declaration. Fields are a superset
// across connector types; the registry interprets them based on Type.
type ConnectorConfig struct {
	Type         string   `yaml:"type"`
	Name         string   `yaml:"name"`
	URL          string   `yaml:"url"`
	PollInterval Duration `yaml:"poll_interval"`
	Brokers      []string `yaml:"brokers"`
	Topic        string   `yaml:"topic"`
	GroupID      string   `yaml:"group_id"`
}

// AIConfig selects and configures the AI provider for the chat plane.
type AIConfig struct {
	Provider  string `yaml:"provider"`
	Model     string `yaml:"model"`
	APIKeyEnv string `yaml:"api_key_env"`
	BaseURL   string `yaml:"base_url"`
}

// Duration is a yaml-friendly time.Duration that parses strings like "5s".
type Duration time.Duration

// UnmarshalYAML parses a duration string (e.g. "5s", "1m").
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	if s == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

// AsDuration returns the value as a time.Duration.
func (d Duration) AsDuration() time.Duration { return time.Duration(d) }

// Load reads and parses configuration from the given path, applying defaults.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return parse(raw)
}

// parse decodes config bytes and applies defaults/validation.
func parse(raw []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}

	for i, c := range cfg.Connectors {
		if c.Name == "" {
			return nil, fmt.Errorf("connector %d: name is required", i)
		}
		if c.Type == "" {
			return nil, fmt.Errorf("connector %q: type is required", c.Name)
		}
	}

	return &cfg, nil
}
