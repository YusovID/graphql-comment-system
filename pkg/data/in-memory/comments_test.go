package inmemory

import (
	"context"
	"graphql-comment-system/graph/model"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInitializeComments(t *testing.T) {
	InitializeComments()

	commentsMutex.RLock()
	defer commentsMutex.RUnlock()

	if len(comments) != 3 {
		t.Fatalf("Expected 3 comments, got %d", len(comments))
	}

	post1Count := 0
	post2Count := 0
	for _, c := range comments {
		switch c.PostID {
		case "1":
			post1Count++
		case "2":
			post2Count++
		}
	}

	if post1Count != 2 {
		t.Errorf("Expected 2 comments for post 1, got %d", post1Count)
	}
	if post2Count != 1 {
		t.Errorf("Expected 1 comment for post 2, got %d", post2Count)
	}
}

func TestGetCommentByID(t *testing.T) {
	InitializeComments()
	ctx := context.Background()

	// Test existing comment
	existingComment := getFirstCommentForPost("1", t)
	comment, err := GetCommentByID(ctx, existingComment.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if comment.ID != existingComment.ID {
		t.Errorf("Expected comment ID %s, got %s", existingComment.ID, comment.ID)
	}

	// Test non-existing comment
	_, err = GetCommentByID(ctx, "non-existing-id")
	if err == nil {
		t.Error("Expected error for non-existing comment, got nil")
	}
}

func TestGetCommentsForPost(t *testing.T) {
	InitializeComments()
	ctx := context.Background()

	// Test basic functionality
	result, err := GetCommentsForPost(ctx, "1", 10, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(result.Comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(result.Comments))
	}

	// Test sorting
	if !isSorted(result.Comments) {
		t.Error("Comments are not sorted by creation time")
	}

	// Test pagination
	firstComment := result.Comments[0]
	paginatedResult, err := GetCommentsForPost(ctx, "1", 1, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(paginatedResult.Comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(paginatedResult.Comments))
	}
	if paginatedResult.Comments[0].ID != firstComment.ID {
		t.Error("Pagination returned wrong comment")
	}
	if !paginatedResult.HasNextPage {
		t.Error("HasNextPage should be true")
	}

	// Test with cursor
	secondResult, err := GetCommentsForPost(ctx, "1", 1, &firstComment.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondResult.Comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(secondResult.Comments))
	}
	if secondResult.Comments[0].ID == firstComment.ID {
		t.Error("Cursor not working properly")
	}
	if secondResult.HasNextPage {
		t.Error("HasNextPage should be false for last comment")
	}

	// Test invalid cursor
	invalidCursor := "invalid"
	invalidResult, err := GetCommentsForPost(ctx, "1", 1, &invalidCursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(invalidResult.Comments) != 1 || invalidResult.Comments[0].ID != firstComment.ID {
		t.Error("Invalid cursor should reset to start")
	}
}

func TestAddComment(t *testing.T) {
	InitializeComments()
	ctx := context.Background()
	initialCount := len(comments)

	newComment := &model.Comment{
		Author:    "Test Author",
		Content:   "Test Content",
		PostID:    "3",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	AddComment(ctx, newComment)

	if len(comments) != initialCount+1 {
		t.Errorf("Expected %d comments, got %d", initialCount+1, len(comments))
	}

	if newComment.ID == "" {
		t.Error("Comment ID was not generated")
	}

	if comment, exists := comments[newComment.ID]; !exists || comment != newComment {
		t.Error("Comment was not properly saved")
	}
}

func TestGetCommentsForPost_ErrorHandling(t *testing.T) {
	InitializeComments()
	ctx := context.Background()

	// Add invalid comment
	invalidComment := &model.Comment{
		ID:        uuid.NewString(),
		Author:    "Invalid",
		Content:   "Invalid",
		PostID:    "1",
		CreatedAt: "invalid-time",
	}
	commentsMutex.Lock()
	comments[invalidComment.ID] = invalidComment
	commentsMutex.Unlock()

	_, err := GetCommentsForPost(ctx, "1", 10, nil)
	if err == nil {
		t.Error("Expected error parsing time, got nil")
	}
}

// Helper functions
func getFirstCommentForPost(postID string, t *testing.T) *model.Comment {
	commentsMutex.RLock()
	defer commentsMutex.RUnlock()

	for _, c := range comments {
		if c.PostID == postID {
			return c
		}
	}
	t.Fatalf("No comments found for post %s", postID)
	return nil
}

func isSorted(comments []*model.Comment) bool {
	return sort.SliceIsSorted(comments, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, comments[i].CreatedAt)
		t2, _ := time.Parse(time.RFC3339, comments[j].CreatedAt)
		return t1.Before(t2)
	})
}