package inmemory

import (
	"context"
	"fmt"
	"graphql-comment-system/graph/model"
	"sort"
	"strconv"
	"sync"
	"time"
)

type CommentStore struct{}

func NewCommentStore() *CommentStore {
	return &CommentStore{}
}

var comments map[string]*model.Comment
var commentsMutex sync.RWMutex

// TODO: Удалить
var commentID int = 3

type CommentsResult struct {
	Comments    []*model.Comment
	HasNextPage bool
}

func InitializeComments() {
	comments = make(map[string]*model.Comment)
	ctx := context.Background()
	store := NewCommentStore()

	comment1 := &model.Comment{
		ID:        "1",
		Author:    "Комментатор 1",
		Content:   "Отличный первый пост!",
		CreatedAt: time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
		PostID:    "1",
	}
	store.AddComment(ctx, comment1)

	comment2 := &model.Comment{
		ID:        "2",
		Author:    "Комментатор 2",
		Content:   "Согласен, очень интересно!",
		CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
		PostID:    "1",
	}
	store.AddComment(ctx, comment2)

	comment3 := &model.Comment{
		ID:        "3",
		Author:    "Комментатор 3",
		Content:   "Мне не очень...",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "2",
	}
	store.AddComment(ctx, comment3)
}

func (*CommentStore) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	commentsMutex.RLock()
	defer commentsMutex.RUnlock()

	comment, ok := comments[id]
	if !ok {
		return nil, fmt.Errorf("comment with id %s not found", id)
	}

	return comment, nil
}

func (*CommentStore) GetCommentsForPost(ctx context.Context, postID string, first int32, afterCursor *string) (*model.CommentConnection, error) {
	commentsMutex.RLock()
	defer commentsMutex.RUnlock()

	filtered := make([]*model.Comment, 0)

	for _, comment := range comments {
		if comment.PostID == postID {
			filtered = append(filtered, comment)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, filtered[i].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, filtered[j].CreatedAt)
		if err1 != nil || err2 != nil {
			return false
		}

		return t1.Before(t2)
	})

	for _, c := range filtered {
		_, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing date: %w", err)
		}
	}

	start := 0

	if afterCursor != nil {
		for i, c := range filtered {
			if c.ID == *afterCursor {
				start = i + 1
				break
			}
		}
	}

	end := start + int(first)
	if end > len(filtered) {
		end = len(filtered)
	}

	commentSlice := filtered[start:end]

	hasNextPage := end < len(filtered)

	return &model.CommentConnection{
		Edges:    convertToCommentEdges(commentSlice),
		PageInfo: &model.PageInfo{HasNextPage: hasNextPage},
	}, nil
}

func convertToCommentEdges(comments []*model.Comment) []*model.CommentEdge {
	edges := make([]*model.CommentEdge, len(comments))

	for i, comment := range comments {
		edges[i] = &model.CommentEdge{
			Node: comment,
		}
	}

	return edges
}

func (*CommentStore) AddComment(ctx context.Context, comment *model.Comment) error {
	commentsMutex.Lock()
	defer commentsMutex.Unlock()

	comments[comment.ID] = comment

	return nil
}

func (*CommentStore) GetRepliesForComment(ctx context.Context, parentID string, first int32, afterCursor *string) (*model.CommentConnection, error) {
	commentsMutex.RLock()
	defer commentsMutex.RUnlock()

	filtered := make([]*model.Comment, 0)

	for _, comment := range comments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			filtered = append(filtered, comment)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, filtered[i].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, filtered[j].CreatedAt)
		if err1 != nil || err2 != nil {
			return false
		}

		return t1.Before(t2)
	})

	for _, c := range filtered {
		_, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing date: %w", err)
		}
	}

	start := 0
	
	if afterCursor != nil {
		for i, c := range filtered {
			if c.ID == *afterCursor {
				start = i + 1
				break
			}
		}
	}

	end := start + int(first)
	if end > len(filtered) {
		end = len(filtered)
	}

	commentSlice := filtered[start:end]

	hasNextPage := end < len(filtered)

	return &model.CommentConnection{
		Edges:    convertToCommentEdges(commentSlice),
		PageInfo: &model.PageInfo{HasNextPage: hasNextPage},
	}, nil
}

// TODO: Удалить
func GetNextCommentID() string {
	ID := strconv.Itoa(commentID)
	commentID++
	return ID
}
