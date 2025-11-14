package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration
type Config struct {
	Mode       string           `toml:"mode"` // "aggregator" or "agent"
	Aggregator AggregatorConfig `toml:"aggregator"`
	Agent      AgentConfig      `toml:"agent"`
}

// AggregatorConfig contains settings for aggregator mode
type AggregatorConfig struct {
	Port     int    `toml:"port"`     // Port to listen on (default 8080)
	Database string `toml:"database"` // SQLite database path
}

// AgentConfig contains settings for agent mode
type AgentConfig struct {
	ListenAddr       string `toml:"listen_addr"`       // Address to listen on (default ":8080")
	AggregatorURL    string `toml:"aggregator_url"`    // URL of the aggregator
	RegisterInterval int    `toml:"register_interval"` // Seconds between registrations (default 300)
}

// LoadConfig loads configuration from a TOML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Set defaults
	if config.Aggregator.Port == 0 {
		config.Aggregator.Port = 8080
	}
	if config.Aggregator.Database == "" {
		config.Aggregator.Database = "sysinfo.db"
	}
	if config.Agent.ListenAddr == "" {
		config.Agent.ListenAddr = ":8080"
	}
	if config.Agent.RegisterInterval == 0 {
		config.Agent.RegisterInterval = 300
	}

	// Validate mode
	if config.Mode != "aggregator" && config.Mode != "agent" {
		return nil, fmt.Errorf("invalid mode: %s (must be 'aggregator' or 'agent')", config.Mode)
	}

	// Validate agent config if in agent mode
	if config.Mode == "agent" && config.Agent.AggregatorURL == "" {
		return nil, fmt.Errorf("aggregator_url is required in agent mode")
	}

	return &config, nil
}

// GenerateDefaultConfig creates a default configuration file
func GenerateDefaultConfig(path string, mode string) error {
	var config Config

	if mode == "aggregator" {
		config = Config{
			Mode: "aggregator",
			Aggregator: AggregatorConfig{
				Port:     8080,
				Database: "sysinfo.db",
			},
		}
	} else {
		config = Config{
			Mode: "agent",
			Agent: AgentConfig{
				ListenAddr:       ":8080",
				AggregatorURL:    "http://localhost:8080",
				RegisterInterval: 300,
			},
		}
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
