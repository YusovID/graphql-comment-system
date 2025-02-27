package main

import (
	"context"
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
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "50051" // Порт по умолчанию для запуска сервера.

func main() {
	// Загрузка переменных окружения из файла .env, если он существует.
	// В противном случае используются системные переменные окружения.
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables: ", err)
	}

	log.Println("Starting server") // Информационное сообщение о старте сервера.

	// Определение порта для сервера.
	// Приоритет отдается переменной окружения PORT, если она задана.
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort // Использование порта по умолчанию, если PORT не задана.
	}

	// Определение типа хранилища данных из переменной окружения STORAGE_TYPE.
	storageType := os.Getenv("STORAGE_TYPE")

	var postStore data.PostStore   // Интерфейс для хранилища постов.
	var commentStore data.CommentStore // Интерфейс для хранилища комментариев.

	// Выбор реализации хранилища данных в зависимости от STORAGE_TYPE.
	switch storageType {
	case "postgres":
		// Конфигурация и инициализация PostgreSQL.
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

		conn, err := postgres.New(config)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer conn.Close(context.Background()) // Гарантированное закрытие соединения после завершения main.

		// Инициализация хранилищ с использованием PostgreSQL.
		postStore = postgres.NewPostStore(conn)
		commentStore = postgres.NewCommentStore(conn)
		log.Println("Using PostgreSQL storage")

	case "inmemory":
		// Инициализация In-Memory хранилища.
		log.Println("Using In-Memory storage")
		inmemory.InitializeData() // Заполнение In-Memory данными по умолчанию.
		postStore = inmemory.NewPostStore()
		commentStore = inmemory.NewCommentStore()

	default:
		// Default case: In-Memory хранилище, если STORAGE_TYPE не задан или не распознан.
		log.Println("STORAGE_TYPE not set or invalid, using default In-Memory storage")
		inmemory.InitializeData() // Заполнение In-Memory данными по умолчанию.
		postStore = inmemory.NewPostStore()
		commentStore = inmemory.NewCommentStore()
	}

	// Создание GraphQL-сервера на основе сгенерированной схемы и resolvers.
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver(postStore, commentStore)}))

	// Добавление транспортов для поддержки различных HTTP-методов и WebSocket.
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Настройка кэширования запросов для повышения производительности.
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Добавление расширений, таких как интроспекция и Automatic Persisted Queries.
	srv.Use(extension.Introspection{}) // Включение интроспекции GraphQL API.
	srv.Use(extension.AutomaticPersistedQuery{ // Поддержка Automatic Persisted Queries для оптимизации запросов.
		Cache: lru.New[string](100),
	})

	// Регистрация HTTP-обработчиков: Playground и основной GraphQL endpoint.
	http.Handle("/", playground.Handler("GraphQL playground", "/query")) // GraphQL Playground для разработки.
	http.Handle("/query", srv)                                         // Основной GraphQL endpoint.

	// Запуск HTTP-сервера и вывод информации в лог.
	log.Printf("Connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil)) // Запуск сервера на заданном порту.
}