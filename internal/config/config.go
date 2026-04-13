package config

import (
	"context"
	"log"
	"time"

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
	// Собираем DSN
	Username string `env:"USER, required"`
	Password string `env:"PASSWORD, required"`
	DBHost   string `env:"HOST, required"`
	DBPort   string `env:"PORT, required"`
	DBName   string `env:"NAME, required"`

    // Параметры SSL (опционально, но полезно)
    SSLMode string `env:"SSL_MODE, default=disable"`

    // Параметры пула соединений
    // MaxConns максимальное количество соединений в пуле (по умолчанию: max(4, runtime.NumCPU()))
    MaxConns int32 `env:"MAX_CONNS, default=25"`
    
    // MinConns минимальное количество соединений в пуле (по умолчанию: 0)
    MinConns int32 `env:"MIN_CONNS, default=5"`
    
    // MaxConnIdleTime максимальное время жизни бездействующего соединения (по умолчанию: 30 минут)
    MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME, default=5m"`
    
    // MaxConnLifetime максимальное время жизни соединения (по умолчанию: 1 час)
    MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME, default=1h"`
    
    // HealthCheckPeriod периодичность проверки здоровья соединений (по умолчанию: 1 минута)
    HealthCheckPeriod time.Duration `env:"HEALTH_CHECK_PERIOD, default=30s"`
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