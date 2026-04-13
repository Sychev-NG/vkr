package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"vkr/internal/config"
)

func New(context context.Context, config *config.Config) *pgx.Conn {
	dbHost := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		config.DB.Username, // Имя пользователя
		config.DB.Password, // Пароль
		config.DB.DBHost,   // Хост базы данных
		config.DB.DBPort,   // Порт базы данных
		config.DB.DBName,   // Название базы данных
	)

	conn, err := pgx.Connect(context, dbHost)
	if err != nil {
		panic(err)
	}

	if err := conn.Ping(context); err != nil {
		panic(err)
	}

	return conn
}