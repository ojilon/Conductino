package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)


type Config struct {
    Window struct{
        Title string `yaml:"title" `
        Width int    `yaml:"width" `
        Height int   `yaml:"height"`
        Debug bool    `yaml:"debug"`
    }`yaml:"window"`
    IPC struct {
        FrontendListen string `yaml:"frontend_listen"`
        BackendURL string `yaml:"backend_url"`
    }`yaml:"ipc"`
    Storage struct {
        DatabasePath string `yaml:"database_path"`
    }`yaml:"storage"`
    Archive struct {
        OutputDir string `yaml:"output_dir"`
        MaxBytes int `yaml:"max_bytes"`
    }`yaml:"archive"`
}

func loadConfig(path string) (*Config, error){
    raw, err := os.ReadFile(path)
    if err != nil {
        return  nil, fmt.Errorf("read config: %w", err)
    }
    var cfg Config
    // gopkg.in/yaml.v3 - strict unmarshal of the YAML tree into the struct.
    if err := yaml.Unmarshal(raw, &cfg); err != nil {
        return  nil, fmt.Errorf("parser config: %w", err)
    }
    return  &cfg, nil
}