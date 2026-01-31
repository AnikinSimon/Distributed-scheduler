package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

// TODO add NATS

type Config struct {
	Env               string        `yaml:"env" env-default:"local"`
	SchedulerInterval time.Duration `yaml:"scheduler_interval" env-default:"1m"`
	NATSURL           string        `yaml:"nats_url"`
	Storage           StorageConfig `yaml:"storage" env-required:"true"`
	HTTP              HTTPConfig    `yaml:"http"`
	Redis             RedisConfig   `yaml:"redis"`
}

type HTTPConfig struct {
	Port int `yaml:"port"`
}

type StorageConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
}

func MustLoad() Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func RedisConnString(cfg RedisConfig) string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}
