package main

import (
	"context"
	"errors"
	"fmt"
	"graphql-comment-system/app/graph"
	"graphql-comment-system/app/pkg/data"
	inmemory "graphql-comment-system/app/pkg/data/in-memory"
	"graphql-comment-system/app/pkg/data/postgres"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "50051"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables: ", err)
	}

	log.Println("Starting server")

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	storageType := os.Getenv("STORAGE_TYPE")

	var postStore data.PostStore
	var commentStore data.CommentStore

	switch storageType {
	case "postgres":
		dbPortStr := os.Getenv("DB_PORT")
		dbPort, err := strconv.Atoi(dbPortStr)
		if err != nil {
			log.Fatalf("Error converting DB_PORT '%s' to int: %v", dbPortStr, err)
		}

		config := postgres.Config{
			Host:     os.Getenv("DB_HOST"),
			Port:     dbPort,
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: os.Getenv("DB_NAME"),
		}

		databaseDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.Username, config.Password, config.Host, config.Port, config.Database)
		serverDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/?sslmode=disable", config.Username, config.Password, config.Host, config.Port, config.Database)

		ctx := context.Background()

		conn, err := postgres.New(config)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "3D000" { // Ошибка "database does not exist"
				log.Println("База данных", config.Database, "не найдена. Попытка создания...")

				serverConn, err := pgx.Connect(ctx, serverDSN)
				if err != nil {
					log.Fatalf("Не удалось подключиться к серверу PostgreSQL для создания базы данных: %v", err)
				}
				defer serverConn.Close(ctx)

				_, err = serverConn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", config.Database))
				if err != nil {
					log.Fatalf("Ошибка при создании базы данных %s: %v", config.Database, err)
				}
				log.Println("База данных", config.Database, "успешно создана.")

				conn, err = postgres.New(config)
				if err != nil {
					log.Fatalf("Не удалось подключиться к созданной базе данных %s: %v", config.Database, err)
				}
				log.Println("Подключение к созданной базе данных", config.Database, "установлено.")

			} else {
				log.Fatalf("Ошибка подключения к базе данных %s: %v", config.Database, err)
			}
		}

		// ** Запуск миграций **
		log.Println("Запуск миграций базы данных...")
		m, err := migrate.New(
			"file://app/internal/migration", // Путь к папке с файлами миграций
			databaseDSN)         // Строка подключения к базе данных
		if err != nil {
			log.Fatalf("Ошибка при создании instance миграций: %v", err)
		}
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) { // Запускаем миграции "вверх"
			log.Fatalf("Ошибка при выполнении миграций: %v", err)
		}
		log.Println("Миграции базы данных успешно выполнены.")

		postStore = postgres.NewPostStore(conn)
		commentStore = postgres.NewCommentStore(conn)
		log.Println("Using PostgreSQL storage")

	case "inmemory":
		log.Println("Using In-Memory storage")
		inmemory.InitializeData()
		postStore = inmemory.NewPostStore()
		commentStore = inmemory.NewCommentStore()

	default:
		log.Println("STORAGE_TYPE not set or invalid, using default In-Memory storage")
		inmemory.InitializeData()
		postStore = inmemory.NewPostStore()
		commentStore = inmemory.NewCommentStore()
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(postStore, commentStore)}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("Connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
