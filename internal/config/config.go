package config

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	API APIConfig
	DB DBConfig `env:", prefix=DB_"`
}

type APIConfig struct {
	Port string `env:"API_PORT, default=8080"`
}

type DBConfig struct {
	Username string `env:"USER, required"`
	Password string `env:"PASSWORD, required"`
	DBHost   string `env:"HOST, required"`
	DBPort   string `env:"PORT, required"`
	DBName   string `env:"NAME, required"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	var cfg Config

	if err := envconfig.Process(ctx, &cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}