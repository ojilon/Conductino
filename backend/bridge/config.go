package bridge

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration structure.
type Config struct {
	Window struct {
		Title string `yaml:"title"`
		Width  int    `yaml:"width"`
		Height int   `yaml:"height"`
		Debug  bool   `yaml:"debug"`
	} `yaml:"window"`
	IPC struct {
		FrontendListen string `yaml:"frontend_listen"`
		BackendURL     string `yaml:"backend_url"`
	} `yaml:"ipc"`
	Storage struct {
		DatabasePath string `yaml:"database_path"`
	} `yaml:"storage"`
	Archive struct {
		OutputDir string `yaml:"output_dir"`
		MaxBytes  int    `yaml:"max_bytes"`
	} `yaml:"archive"`
}

// loadConfig loads and parses the YAML configuration file.
func loadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parser config: %w", err)
	}
	return &cfg, nil
}