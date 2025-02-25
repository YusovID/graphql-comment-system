// cmd/main.go
package main

import (
	"graphql-comment-system/graph"
	inmemoryData "graphql-comment-system/pkg/data/in-memory"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func main() {
	// Загрузка .env файла
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	storageType := os.Getenv("STORAGE_TYPE")

	var postStore inmemoryData.PostStore
	var commentStore inmemoryData.CommentStore

	inMemoryStore := &inmemoryData.InMemoryStore{} // Create InMemoryStore instance

	switch storageType {
	case "postgres":
		// TODO: Инициализация PostgreSQL хранилища (реализация будет позже)
		log.Println("Using PostgreSQL storage (not fully implemented yet)")
		// Временно используем in-memory как заглушку for postgres - now use InMemoryStore directly
		inmemoryData.InitializeData() // Keep InitializeData for initial data setup
		postStore = inMemoryStore
		commentStore = inMemoryStore

	case "inmemory":
		log.Println("Using In-Memory storage")
		inmemoryData.InitializeData() // Keep InitializeData for initial data setup
		postStore = inMemoryStore
		commentStore = inMemoryStore

	default:
		log.Printf("STORAGE_TYPE not set or invalid, using default In-Memory storage")
		inmemoryData.InitializeData() // Keep InitializeData for initial data setup
		postStore = inMemoryStore
		commentStore = inMemoryStore
	}

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		PostStore:    postStore,
		CommentStore: commentStore,
	}}))

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
