package graph

import (
	"graphql-comment-system/pkg/data"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	PostStore    data.PostStore
	CommentStore data.CommentStore
}

func NewResolver(postStore data.PostStore, commentStore data.CommentStore) *Resolver {
	return &Resolver{PostStore: postStore, CommentStore: commentStore}
}
