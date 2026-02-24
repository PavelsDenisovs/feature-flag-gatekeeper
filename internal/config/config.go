package config

type Config struct {
	Port int
	DBUrl string
}

func LoadConfig(configFilePath string) (Config, error) {
	return Config{}, nil
}