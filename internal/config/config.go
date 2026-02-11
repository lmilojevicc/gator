package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "config.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getDefaults() Config {
	return Config{
		DBURL:           "postgres://localhost:5432/gator?sslmode=disable",
		CurrentUserName: "",
	}
}

func (cfg *Config) SetUser(name string) error {
	cfg.CurrentUserName = name
	return cfg.Save()
}

func (cfg *Config) Save() error {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("getting config filepath: %w", err)
	}

	configDir := filepath.Dir(configFilePath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	err = os.WriteFile(configFilePath, data, 0o600)
	if err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func getConfigFilePath() (string, error) {
	switch {
	case os.Getenv("GATOR_CONFIG") != "":
		return os.Getenv("GATOR_CONFIG"), nil
	case os.Getenv("XDG_CONFIG_HOME") != "":
		return filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "gator", configFileName), nil
	default:
		home := os.Getenv("HOME")
		if home == "" {
			return "", fmt.Errorf("HOME environment variable not set")
		}
		return filepath.Join(home, ".config", "gator", configFileName), nil
	}
}

func Read() (Config, error) {
	cfg := getDefaults()
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("%v", err)
	}

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading config file: %v", err)
	}

	if err = json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %v", err)
	}

	return cfg, nil
}
