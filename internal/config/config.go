package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddress string       `yaml:"listen_address" json:"listen_address" toml:"listen_address"`
	UseIOUring    bool         `yaml:"use_iouring" json:"use_iouring" toml:"use_iouring"`
	Algorithm     string       `yaml:"algorithm" json:"algorithm" toml:"algorithm"`
	Backends       []BackendCfg `yaml:"backends" json:"backends" toml:"backends"`
	HealthCheck   HealthCfg    `yaml:"health_check" json:"health_check" toml:"health_check"`
	Timeout       TimeoutCfg   `yaml:"timeout" json:"timeout" toml:"timeout"`
}

type BackendCfg struct {
	Address string `yaml:"address" json:"address" toml:"address"`
	Weight  int    `yaml:"weight" json:"weight" toml:"weight"`
}

type HealthCfg struct {
	IntervalSec int `yaml:"interval_sec" json:"interval_sec" toml:"interval_sec"`
	TimeoutSec  int `yaml:"timeout_sec" json:"timeout_sec" toml:"timeout_sec"`
	Retries     int `yaml:"retries" json:"retries" toml:"retries"`
}

type TimeoutCfg struct {
	ClientIdleSec  int `yaml:"client_idle_sec" json:"client_idle_sec" toml:"client_idle_sec"`
	BackendIdleSec int `yaml:"backend_idle_sec" json:"backend_idle_sec" toml:"backend_idle_sec"`
	ConnectTimeout int `yaml:"connect_timeout" json:"connect_timeout" toml:"connect_timeout"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)

	if err != nil {
		return nil, fmt.Errorf("failed to read the file: %w", err)
	}

	cfg := &Config{}
	ext := filepath.Ext(filename)

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON %w", err)
		}
	case ".toml":
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse TOML %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format %s", ext)
	}

	cfg.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.ListenAddress == "" {
		return errors.New("listen Address is empty")
	}

	if len(c.Backends) == 0 {
		return errors.New("no Backends are specified")
	}

	validModes := map[string]bool{
		"round_robin":       true,
		"least_connections": true,
		"weighted":          true,
		"ip_hash":           true,
	}

	if !validModes[c.Algorithm] {
		return errors.New("invalid load balancing algorithm")
	}

	return nil
}

func (c *Config) applyDefaults() {

	if c.Algorithm == "" {
		c.Algorithm = "round_robin"
	}
	if c.HealthCheck.IntervalSec == 0 {
		c.HealthCheck.IntervalSec = 5
	}
	if c.HealthCheck.TimeoutSec == 0 {
		c.HealthCheck.TimeoutSec = 2
	}
	if c.HealthCheck.Retries == 0 {
		c.HealthCheck.Retries = 2
	}
	if c.Timeout.ClientIdleSec == 0 {
		c.Timeout.ClientIdleSec = 30
	}
	if c.Timeout.BackendIdleSec == 0 {
		c.Timeout.BackendIdleSec = 30
	}
	if c.Timeout.ConnectTimeout == 0 {
		c.Timeout.ConnectTimeout = 3
	}
}
