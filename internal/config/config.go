package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port int
	DBUrl string
}

func LoadConfig(configFilePath string) (Config, error) {
	if configFilePath != "" {
		if err := godotenv.Load(configFilePath); err != nil {
			return Config{}, fmt.Errorf("failed to load config file by path %s: %w", configFilePath, err)
		}
	} else {
		godotenv.Load()
	}
  
	cfg, err := loadEnvConfig()
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse environment variables into config: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func loadEnvConfig() (Config, error) {
	port, err := getIntEnv("PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:  port,
		DBUrl: os.Getenv("DB_URL"),
	}, nil
}

func validateConfig(cfg Config) error {
	var errs []error

	if cfg.DBUrl == "" {
		errs = append(errs, fmt.Errorf("required environment variable DB_URL is not set"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %v", errs)
	}

	return nil
}

// returns 0 if not set
func getIntEnv(key string, defaultVal int) (int, error) {
	val := os.Getenv(key) 
	if val == "" {
		return defaultVal, nil
	}
	return strconv.Atoi(val)
}