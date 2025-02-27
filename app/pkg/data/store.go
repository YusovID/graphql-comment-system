// pkg/data/store.go
package data

import (
	"context"
	"graphql-comment-system/graph/model"
)

// PostStore interface defines methods for interacting with post data.
type PostStore interface {
	AddPost(ctx context.Context, post *model.Post) error
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	GetPosts(ctx context.Context, first int32, after *string) (*model.PostConnection, error)
}

// CommentStore interface defines methods for interacting with comment data.
type CommentStore interface {
	AddComment(ctx context.Context, comment *model.Comment) error
	GetCommentByID(ctx context.Context, id string) (*model.Comment, error)
	GetCommentsForPost(ctx context.Context, postID string, first int32, after *string) (*model.CommentConnection, error)
	GetRepliesForComment(ctx context.Context, commentID string, first int32, after *string) (*model.CommentConnection, error)
}
