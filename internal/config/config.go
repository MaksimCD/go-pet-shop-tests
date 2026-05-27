package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env             string        `yaml:"env" env-default:"local"`
	DatabaseURL     string        `env:"DATABASE_URL"`
	DatabaseTimeout time.Duration `yaml:"database_timeout" env:"DB_TIMEOUT" env-default:"10s"`
	HTTPServer      HTTPServer    `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	_ = godotenv.Load()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	configPath := filepath.Join("config", fmt.Sprintf("%s.yaml", env))
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file not found: %s", configPath))
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("cannot read environment: " + err.Error())
	}
	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL not set in environment")
	}

	return &cfg
}
