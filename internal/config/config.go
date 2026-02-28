// Package config provides persistent configuration for the prayer-times CLI.
//
// Configuration is stored as JSON at ~/.config/prayer-times/config.json
// (XDG-compliant). The merge priority is: CLI flags > config file > defaults.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	configDirName  = "prayer-times"
	configFileName = "config.json"
)

// ValidKeys lists all config keys that can be set via `config set`.
var ValidKeys = []string{
	"city", "country",
	"latitude", "longitude",
	"method", "school",
	"time_format",
	"prayers",
	"cache_dir",
}

// Config holds all user-configurable settings.
// Zero values mean "not set" (use defaults or auto-detect).
type Config struct {
	City       string  `json:"city,omitempty"`
	Country    string  `json:"country,omitempty"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
	Method     *int    `json:"method,omitempty"`      // pointer so we can distinguish "not set" from 0
	School     *int    `json:"school,omitempty"`      // pointer so we can distinguish "not set" from 0
	TimeFormat string  `json:"time_format,omitempty"` // "12h" or "24h"
	Prayers    string  `json:"prayers,omitempty"`     // comma-separated list
	CacheDir   string  `json:"cache_dir,omitempty"`
}

// Defaults returns a Config with all default values applied.
func Defaults() Config {
	method := -1
	school := -1
	return Config{
		Method:     &method,
		School:     &school,
		TimeFormat: "24h",
	}
}

// Dir returns the config directory path.
// It respects $XDG_CONFIG_HOME if set, otherwise uses ~/.config/.
func Dir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, configDirName), nil
}

// Path returns the full path to the config file.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load reads the config file from disk.
// If the file does not exist, it returns an empty Config (not an error).
// If the file exists but is invalid JSON, it returns an error.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	return LoadFrom(path)
}

// LoadFrom reads the config from a specific file path.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg := Config{}
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file %s: %w", path, err)
	}

	return &cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}

	return c.SaveTo(path)
}

// SaveTo writes the config to a specific file path.
func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Reset deletes the config file.
func Reset() error {
	path, err := Path()
	if err != nil {
		return err
	}

	return ResetAt(path)
}

// ResetAt deletes the config file at a specific path.
func ResetAt(path string) error {
	err := os.Remove(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to delete config file: %w", err)
	}
	return nil
}

// Set sets a config key to the given value.
// It validates the key name and parses the value into the correct type.
func (c *Config) Set(key, value string) error {
	switch key {
	case "city":
		c.City = value
	case "country":
		c.Country = value
	case "latitude":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid latitude %q: must be a number", value)
		}
		if v < -90 || v > 90 {
			return fmt.Errorf("invalid latitude %q: must be between -90 and 90", value)
		}
		c.Latitude = v
	case "longitude":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid longitude %q: must be a number", value)
		}
		if v < -180 || v > 180 {
			return fmt.Errorf("invalid longitude %q: must be between -180 and 180", value)
		}
		c.Longitude = v
	case "method":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid method %q: must be an integer", value)
		}
		if v < 0 || v > 23 {
			return fmt.Errorf("invalid method %q: must be between 0 and 23", value)
		}
		c.Method = &v
	case "school":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid school %q: must be an integer", value)
		}
		if v != 0 && v != 1 {
			return fmt.Errorf("invalid school %q: must be 0 (Shafi) or 1 (Hanafi)", value)
		}
		c.School = &v
	case "time_format":
		if value != "12h" && value != "24h" {
			return fmt.Errorf("invalid time_format %q: must be \"12h\" or \"24h\"", value)
		}
		c.TimeFormat = value
	case "prayers":
		// Validate each prayer name.
		names := strings.Split(value, ",")
		for _, n := range names {
			n = strings.TrimSpace(n)
			if !isValidPrayerName(n) {
				return fmt.Errorf("invalid prayer name %q in prayers list", n)
			}
		}
		c.Prayers = value
	case "cache_dir":
		c.CacheDir = value
	default:
		return fmt.Errorf("unknown config key %q; valid keys: %s", key, strings.Join(ValidKeys, ", "))
	}

	return nil
}

// Get returns the string value of a config key.
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "city":
		return c.City, nil
	case "country":
		return c.Country, nil
	case "latitude":
		if c.Latitude == 0 {
			return "", nil
		}
		return strconv.FormatFloat(c.Latitude, 'f', -1, 64), nil
	case "longitude":
		if c.Longitude == 0 {
			return "", nil
		}
		return strconv.FormatFloat(c.Longitude, 'f', -1, 64), nil
	case "method":
		if c.Method == nil {
			return "", nil
		}
		return strconv.Itoa(*c.Method), nil
	case "school":
		if c.School == nil {
			return "", nil
		}
		return strconv.Itoa(*c.School), nil
	case "time_format":
		return c.TimeFormat, nil
	case "prayers":
		return c.Prayers, nil
	case "cache_dir":
		return c.CacheDir, nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

// validPrayerNames are the prayer names the API supports.
var validPrayerNames = map[string]bool{
	"Fajr": true, "Sunrise": true, "Dhuhr": true, "Asr": true,
	"Sunset": true, "Maghrib": true, "Isha": true,
	"Imsak": true, "Midnight": true, "Firstthird": true, "Lastthird": true,
}

func isValidPrayerName(name string) bool {
	return validPrayerNames[name]
}

// MethodOrDefault returns the method value, falling back to the given default.
func (c *Config) MethodOrDefault(def int) int {
	if c.Method != nil {
		return *c.Method
	}
	return def
}

// SchoolOrDefault returns the school value, falling back to the given default.
func (c *Config) SchoolOrDefault(def int) int {
	if c.School != nil {
		return *c.School
	}
	return def
}
