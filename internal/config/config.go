package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port  int
	DBURL string
}

type Overrides struct {
	Port  *int
	DBURL *string
}

// Apply overrides cfg with provided fields
func (cfg *Config) Apply(overrides Overrides) {
	if overrides.Port != nil {
		cfg.Port = *overrides.Port
	}
	if overrides.DBURL != nil {
		cfg.DBURL = *overrides.DBURL
	}
}

// Load loads configuration from environment after loading a configuration
// from config file by path.
func Load(path string) (Config, error) {
	if path != "" {
		if err := godotenv.Load(path); err != nil {
			return Config{}, fmt.Errorf("failed to load config file by path %s: %w", path, err)
		}
	} else {
		_ = godotenv.Load()
	}

	cfg, err := fromEnv()
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config from environment: %w", err)
	}

	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func fromEnv() (Config, error) {
	port, err := getIntEnv("PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:  port,
		DBURL: os.Getenv("DB_URL"),
	}, nil
}

func validate(cfg Config) error {
	var errs []error

	if cfg.DBURL == "" {
		errs = append(errs, fmt.Errorf("required environment variable DB_URL is not set"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %v", errors.Join(errs...))
	}

	return nil
}

// returns defaultVal if not set
func getIntEnv(key string, defaultVal int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(val)
}
