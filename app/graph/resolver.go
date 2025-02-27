package graph

import (
	"graphql-comment-system/app/pkg/data"
)

// Resolver - структура для хранения зависимостей, необходимых для resolvers GraphQL.
// Используется для dependency injection в приложении.
type Resolver struct {
	PostStore    data.PostStore    // Интерфейс для доступа к данным постов.
	CommentStore data.CommentStore // Интерфейс для доступа к данным комментариев.
}

// NewResolver - конструктор для создания экземпляра Resolver.
// Принимает реализации интерфейсов PostStore и CommentStore и возвращает Resolver,
// готовый к использованию в resolvers GraphQL.
func NewResolver(postStore data.PostStore, commentStore data.CommentStore) *Resolver {
	return &Resolver{PostStore: postStore, CommentStore: commentStore}
}
