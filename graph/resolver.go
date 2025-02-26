// graph/resolver.go
package graph

import "graphql-comment-system/pkg/data/in-memory"

// Resolver struct holds the application state (dependencies).
type Resolver struct {
	PostStore    inmemory.PostStore
	CommentStore inmemory.CommentStore
}

// NewResolver ...
func NewResolver(postStore inmemory.PostStore, commentStore inmemory.CommentStore) *Resolver {
	return &Resolver{PostStore: postStore, CommentStore: commentStore}
}


