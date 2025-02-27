package main

import (
	"graphql-comment-system/graph"
	"graphql-comment-system/pkg/data"
	inmemoryData "graphql-comment-system/pkg/data/in-memory"
	"graphql-comment-system/pkg/data/postgres"
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

const defaultPort = "50051"

func main() {
	// Загрузка .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables: ", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	storageType := os.Getenv("STORAGE_TYPE")

	var postStore data.PostStore
	var commentStore data.CommentStore

	switch storageType {
	case "postgres":
		port, err := strconv.Atoi(os.Getenv("DB_PORT"))
		if err != nil {
			log.Fatalf("Error converting DB_PORT to int: %v", err)
		}

		config := postgres.Config{
			Host:     os.Getenv("DB_HOST"),
			Port:     port,
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: os.Getenv("DB_NAME"),
		}
		conn, err := postgres.New(config)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		postStore = postgres.NewPostStore(conn)
		commentStore = postgres.NewCommentStore(conn)
		log.Println("Using PostgreSQL storage")

	case "inmemory":
		log.Println("Using In-Memory storage")
		postStore = inmemoryData.NewPostStore()
		commentStore = inmemoryData.NewCommentStore()

	default:
		log.Printf("STORAGE_TYPE not set or invalid, using default In-Memory storage")
		postStore = inmemoryData.NewPostStore()
		commentStore = inmemoryData.NewCommentStore()
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

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
