//go:build windows

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	LockMinutes int  `json:"lock_minutes"`
	Enabled     bool `json:"enabled"`
}

func defaultConfig() Config {
	return Config{
		LockMinutes: 15,
		Enabled:     true,
	}
}

func defaultConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func loadOrCreateConfig(path string) Config {
	cfg, err := loadConfig(path)
	if err == nil {
		return cfg
	}
	log.Printf("load config failed, using defaults: %v", err)
	return defaultConfig()
}

func loadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.LockMinutes <= 0 {
		cfg.LockMinutes = defaultConfig().LockMinutes
	}
	return cfg, nil
}

func saveConfig(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
