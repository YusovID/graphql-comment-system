package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

func New(config Config) (*pgx.Conn, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.Username, config.Password, config.Host, config.Port, config.Database)
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return conn, nil
}

// PostStore struct
type PostStore struct {
	conn *pgx.Conn
}

// NewPostStore returns new instance of PostStore
func NewPostStore(*pgx.Conn) *PostStore {
	return &PostStore{}
}

// CommentStore struct
type CommentStore struct {
	conn *pgx.Conn
}

// NewCommentStore returns new instance of CommentStore
func NewCommentStore(*pgx.Conn) *CommentStore {
	return &CommentStore{}
}
