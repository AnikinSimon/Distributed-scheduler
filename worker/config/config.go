package config

import (
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	NATSURL string `yaml:"nats_url"`
}

func MustLoad() *Config {
	path := fetchPath()

	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); err != nil {
		panic("config file not found: " + path)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}

func fetchPath() string {
	path := os.Getenv("CONFIG_PATH")
	return path
}
