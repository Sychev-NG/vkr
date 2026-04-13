package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"vkr/internal/config"
)


func NewPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	dbHost := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		cfg.DB.Username, // Имя пользователя
		cfg.DB.Password, // Пароль
		cfg.DB.DBHost,   // Хост базы данных
		cfg.DB.DBPort,   // Порт базы данных
		cfg.DB.DBName,   // Название базы данных
	)

    // 1. Парсим строку подключения в стандартную конфигурацию pgx
    poolConfig, err := pgxpool.ParseConfig(dbHost)
    if err != nil {
        return nil, fmt.Errorf("failed to parse DSN: %w", err)
    }

    // 2. Применяем пользовательские настройки пула (если они заданы)
    if cfg.DB.MaxConns > 0 {
        poolConfig.MaxConns = cfg.DB.MaxConns
    }
    if cfg.DB.MinConns > 0 {
        poolConfig.MinConns = cfg.DB.MinConns
    }
    if cfg.DB.MaxConnIdleTime > 0 {
        poolConfig.MaxConnIdleTime = cfg.DB.MaxConnIdleTime
    }
    if cfg.DB.MaxConnLifetime > 0 {
        poolConfig.MaxConnLifetime = cfg.DB.MaxConnLifetime
    }
    if cfg.DB.HealthCheckPeriod > 0 {
        poolConfig.HealthCheckPeriod = cfg.DB.HealthCheckPeriod
    }

    // 3. Создаем пул соединений
    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    // 4. Проверяем, что соединение с БД действительно устанавливается
    if err := pool.Ping(ctx); err != nil {
        // Если пинг не удался, закрываем пул, чтобы не оставлять висящие ресурсы
        pool.Close()
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return pool, nil
}