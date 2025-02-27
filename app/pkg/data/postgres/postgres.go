package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Config - структура конфигурации для подключения к PostgreSQL.
type Config struct {
	Host     string // Host - адрес хоста базы данных.
	Port     int    // Port - порт для подключения к базе данных.
	Username string // Username - имя пользователя для аутентификации в базе данных.
	Password string // Password - пароль пользователя для аутентификации.
	Database string // Database - имя базы данных, к которой необходимо подключиться.
}

// New - функция создания нового подключения к базе данных PostgreSQL.
func New(config Config) (*pgx.Conn, error) {
	// Формирование строки подключения к базе данных на основе параметров конфигурации.
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.Username, config.Password, config.Host, config.Port, config.Database)
	// Установление соединения с базой данных PostgreSQL с использованием сформированной строки подключения.
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		// В случае ошибки подключения, возвращается nil и форматированная ошибка.
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return conn, nil // В случае успешного подключения, возвращается экземпляр pgx.Conn и nil (отсутствие ошибки).
}
