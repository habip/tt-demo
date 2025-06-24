package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-default:"local"`
	HTTPServer HTTPServer `yaml:"http_server"`
	DB         DB         `yaml:"db"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080" env:"HTTP_SERVER_ADDRESS"`
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

type DB struct {
	Host     string `yaml:"host" env-required:"true" env:"DB_HOST"`
	Port     int    `yaml:"port" env-required:"true" env:"DB_PORT"`
	Username string `yaml:"username" env-required:"true" env:"DB_USERNAME"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file not exists %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %s", err)
	}

	return &cfg
}
