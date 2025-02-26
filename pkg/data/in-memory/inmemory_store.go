// pkg/data/in-memory/store.go
package inmemory

import (
	"context"
	"graphql-comment-system/graph/model"
)

// PostStore interface for post data access methods
type PostStore interface {
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	GetPosts(ctx context.Context, first int32, afterCursor *string) (PostsResult, error)
	AddPost(ctx context.Context, post *model.Post)
}

// CommentStore interface for comment data access methods
type CommentStore interface {
	GetCommentByID(ctx context.Context, id string) (*model.Comment, error)
	GetCommentsForPost(ctx context.Context, postID string, first int32, afterCursor *string) (CommentsResult, error)
	AddComment(ctx context.Context, comment *model.Comment)
	GetRepliesForComment(ctx context.Context, commentID string, first int32, afterCursor *string) (CommentsResult, error)
}

// InMemoryStore implements PostStore and CommentStore using in-memory data
type InMemoryStore struct {
	// You can embed your existing in-memory data structures here if needed,
	// or manage them directly within the methods of InMemoryStore.
	// For now, we'll rely on the existing `posts` and `comments` maps
}

// Ensure InMemoryStore actually implements the interfaces at compile time
var _ PostStore = &InMemoryStore{}
var _ CommentStore = &InMemoryStore{}

// Implement PostStore methods for InMemoryStore

func (s *InMemoryStore) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	return GetPostByID(ctx, id) // Directly call existing functions
}

func (s *InMemoryStore) GetPosts(ctx context.Context, first int32, afterCursor *string) (PostsResult, error) {
	return GetPosts(ctx, first, afterCursor) // Directly call existing functions
}

func (s *InMemoryStore) AddPost(ctx context.Context, post *model.Post) {
	AddPost(ctx, post) // Directly call existing functions
}

// Implement CommentStore methods for InMemoryStore

func (s *InMemoryStore) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	return GetCommentByID(ctx, id) // Assuming you have GetCommentByID in comments.go
}

func (s *InMemoryStore) GetCommentsForPost(ctx context.Context, postID string, first int32, afterCursor *string) (CommentsResult, error) {
	return GetCommentsForPost(ctx, postID, first, afterCursor) // Directly call existing functions
}

func (s *InMemoryStore) AddComment(ctx context.Context, comment *model.Comment) {
	AddComment(ctx, comment) // Directly call existing functions
}

func (s *InMemoryStore) GetRepliesForComment(ctx context.Context, commentID string, first int32, afterCursor *string) (CommentsResult, error) {
	return GetRepliesForComment(ctx, commentID, first, afterCursor) // Directly call existing functions
}
