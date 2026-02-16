package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName  = ".config/loto-cli"
	configFileName = "config.json"
)

// DefaultUserAgent is the default user agent string
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"

// Config holds the user credentials for bilete.loto.ro
type Config struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	UserAgent string `json:"user_agent"`
}

// ErrCredentialsMissing is returned when email or password is empty
var ErrCredentialsMissing = errors.New("credentials missing: please set email and password in config file")

// GetConfigDir returns the full path to the config directory
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, configDirName), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

// EnsureExists creates the config directory and a prefilled config file if they don't exist.
// Returns true if a new config file was created (meaning user needs to fill it in).
func EnsureExists() (bool, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return false, err
	}

	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return false, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		emptyConfig := Config{
			Email:     "",
			Password:  "",
			UserAgent: DefaultUserAgent,
		}
		data, err := json.MarshalIndent(emptyConfig, "", "  ")
		if err != nil {
			return false, err
		}
		if err := os.WriteFile(configPath, data, 0600); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// Load reads and parses the config file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in config file: %w\nPlease check the syntax at: %s", err, configPath)
	}

	if cfg.Email == "" || cfg.Password == "" {
		return nil, ErrCredentialsMissing
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}

	return &cfg, nil
}
