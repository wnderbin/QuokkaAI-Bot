package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	TgToken       string `yaml:"telegram-token"`
	DeepSeekToken string `yaml:"deepseek-token"`
	BaseURL       string `yaml:"base-url"`
	DeepSeekModel string `yaml:"deepseek-model"`
	Debug         bool   `yaml:"debug-mode"`
}

var AES_KEY string // message encryption key

func Load() *Config {
	config_path := os.Getenv("CONFIG_PATH")
	if config_path == "" {
		log.Fatal("[ config.go ] CONFIG_PATH is not set")
	}
	if _, err := os.Stat(config_path); os.IsExist(err) {
		log.Fatalf("[ config.go ] Config is not exist: %s\n", config_path)
	}
	var conf Config

	if err := cleanenv.ReadConfig(config_path, &conf); err != nil {
		log.Fatalf("[ config.go ] Cannot read config: %s\n", config_path)
	}

	AES_KEY = os.Getenv("AES_KEY")
	if AES_KEY == "" {
		log.Fatal("[ config.go ] AES-KEY is not set")
	}

	return &conf
}
