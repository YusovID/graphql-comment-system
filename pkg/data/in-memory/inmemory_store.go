// pkg/data/in-memory/inmemory_store.go
package inmemory

import (
	"context"
	"graphql-comment-system/graph/model"
	"graphql-comment-system/pkg/data"
)

// InMemoryStore implements PostStore and CommentStore using in-memory data
type InMemoryStore struct{}

// Ensure InMemoryStore actually implements the interfaces at compile time
var _ data.PostStore = &InMemoryStore{}
var _ data.CommentStore = &InMemoryStore{}

// Implement PostStore methods for InMemoryStore

func (s *InMemoryStore) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	return GetPostByID(ctx, id)
}

func (s *InMemoryStore) GetPosts(ctx context.Context, first int32, afterCursor *string) (*model.PostConnection, error) {
	result, err := GetPosts(ctx, first, afterCursor)
	if err != nil {
		return nil, err
	}

	var postEdges []*model.PostEdge
	for _, post := range result.Posts {
		postEdges = append(postEdges, &model.PostEdge{
			Node:   post,
			Cursor: post.ID,
		})
	}
	pageInfo := &model.PageInfo{
		HasNextPage: result.HasNextPage,
	}
	return &model.PostConnection{
		Edges:    postEdges,
		PageInfo: pageInfo,
	}, nil
}

func (s *InMemoryStore) AddPost(ctx context.Context, post *model.Post) error {
	AddPost(ctx, post)
	return nil
}

// Implement CommentStore methods for InMemoryStore

func (s *InMemoryStore) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	return GetCommentByID(ctx, id)
}

func (s *InMemoryStore) GetCommentsForPost(ctx context.Context, postID string, first int32, afterCursor *string) (*model.CommentConnection, error) {
	result, err := GetCommentsForPost(ctx, postID, first, afterCursor)
	if err != nil {
		return nil, err
	}

	var commentEdges []*model.CommentEdge
	for _, comment := range result.Comments {
		commentEdges = append(commentEdges, &model.CommentEdge{
			Node:   comment,
			Cursor: comment.ID,
		})
	}
	pageInfo := &model.PageInfo{
		HasNextPage: result.HasNextPage,
	}
	return &model.CommentConnection{
		Edges:    commentEdges,
		PageInfo: pageInfo,
	}, nil
}

func (s *InMemoryStore) AddComment(ctx context.Context, comment *model.Comment) error {
	AddComment(ctx, comment)
	return nil
}

func (s *InMemoryStore) GetRepliesForComment(ctx context.Context, commentID string, first int32, afterCursor *string) (*model.CommentConnection, error) {
	result, err := GetRepliesForComment(ctx, commentID, first, afterCursor)
	if err != nil {
		return nil, err
	}

	var commentEdges []*model.CommentEdge
	for _, comment := range result.Comments {
		commentEdges = append(commentEdges, &model.CommentEdge{
			Node:   comment,
			Cursor: comment.ID,
		})
	}
	pageInfo := &model.PageInfo{
		HasNextPage: result.HasNextPage,
	}
	return &model.CommentConnection{
		Edges:    commentEdges,
		PageInfo: pageInfo,
	}, nil
}

func NewPostStore() data.PostStore {
	return &InMemoryStore{}
}

func NewCommentStore() data.CommentStore {
	return &InMemoryStore{}
}
