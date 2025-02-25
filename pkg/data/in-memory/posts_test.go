package inmemory

import (
	"context"
	"fmt"
	"graphql-comment-system/graph/model"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInitializePosts(t *testing.T) {
	InitializePosts()

	postsMutex.RLock()
	defer postsMutex.RUnlock()

	if len(posts) != 2 {
		t.Fatalf("Expected 2 posts, got %d", len(posts))
	}

	for _, post := range posts {
		switch post.Title {
		case "Первый пост":
			if !post.AllowComments {
				t.Error("Post 1 should allow comments")
			}
		case "Второй пост":
			if post.AllowComments {
				t.Error("Post 2 should not allow comments")
			}
		default:
			t.Errorf("Unexpected post title: %s", post.Title)
		}
	}
}

func TestGetPostByID(t *testing.T) {
	InitializePosts()
	ctx := context.Background()

	// Get existing post
	existingPost := getFirstPost(t)
	post, err := GetPostByID(ctx, existingPost.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if post.ID != existingPost.ID {
		t.Errorf("Expected post ID %s, got %s", existingPost.ID, post.ID)
	}

	// Get non-existing post
	_, err = GetPostByID(ctx, "non-existing-id")
	if err == nil {
		t.Error("Expected error for non-existing post, got nil")
	}
}

func TestGetPosts(t *testing.T) {
	InitializePosts()
	ctx := context.Background()
	addTestPosts(ctx, t)

	// Test basic functionality
	result, err := GetPosts(ctx, 10, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test sorting order (newest first)
	if !isPostsSorted(result.Posts) {
		t.Error("Posts are not sorted by creation time (newest first)")
	}

	// Test pagination
	firstPost := result.Posts[0]
	paginatedResult, err := GetPosts(ctx, 1, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(paginatedResult.Posts) != 1 {
		t.Fatalf("Expected 1 post, got %d", len(paginatedResult.Posts))
	}
	if paginatedResult.Posts[0].ID != firstPost.ID {
		t.Error("Pagination returned wrong post")
	}
	if !paginatedResult.HasNextPage {
		t.Error("HasNextPage should be true")
	}

	// Test cursor pagination
	secondResult, err := GetPosts(ctx, 1, &firstPost.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondResult.Posts) != 1 {
		t.Fatalf("Expected 1 post, got %d", len(secondResult.Posts))
	}
	if secondResult.Posts[0].ID == firstPost.ID {
		t.Error("Cursor not working properly")
	}

	// Test invalid cursor
	invalidCursor := "invalid"
	invalidResult, err := GetPosts(ctx, 10, &invalidCursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(invalidResult.Posts) != len(result.Posts) {
		t.Error("Invalid cursor should return all posts")
	}
}

func TestGetPostsWithInvalidDates(t *testing.T) {
	InitializePosts()
	ctx := context.Background()

	// Add post with invalid date
	invalidPost := &model.Post{
		ID:        uuid.NewString(),
		Title:     "Invalid Date Post",
		Content:   "Invalid content",
		Author:    "Tester",
		CreatedAt: "invalid-time-format",
	}
	postsMutex.Lock()
	posts[invalidPost.ID] = invalidPost
	postsMutex.Unlock()

	result, err := GetPosts(ctx, 10, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if invalid post is last
	lastPost := result.Posts[len(result.Posts)-1]
	if lastPost.ID != invalidPost.ID {
		t.Error("Post with invalid date should be last")
	}
}

func TestAddPost(t *testing.T) {
	InitializePosts()
	ctx := context.Background()
	initialCount := len(posts)

	newPost := &model.Post{
		Title:         "New Post",
		Content:       "New Content",
		Author:        "New Author",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: true,
	}

	AddPost(ctx, newPost)

	postsMutex.RLock()
	defer postsMutex.RUnlock()

	if len(posts) != initialCount+1 {
		t.Errorf("Expected %d posts, got %d", initialCount+1, len(posts))
	}

	if newPost.ID == "" {
		t.Error("Post ID was not generated")
	}

	if post, exists := posts[newPost.ID]; !exists || post != newPost {
		t.Error("Post was not properly saved")
	}
}

func TestGetPostsEdgeCases(t *testing.T) {
	InitializePosts()
	ctx := context.Background()

	// Test first=0
	result, err := GetPosts(ctx, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Posts) != len(posts) {
		t.Errorf("Expected all %d posts, got %d", len(posts), len(result.Posts))
	}

	// Test after last post
	lastPostID := getLastPostID(t)
	emptyResult, err := GetPosts(ctx, 10, &lastPostID)
	if err != nil {
		t.Fatal(err)
	}
	if len(emptyResult.Posts) != 0 {
		t.Errorf("Expected 0 posts, got %d", len(emptyResult.Posts))
	}
	if emptyResult.HasNextPage {
		t.Error("HasNextPage should be false")
	}
}

// Helpers
func getFirstPost(t *testing.T) *model.Post {
	postsMutex.RLock()
	defer postsMutex.RUnlock()

	for _, post := range posts {
		return post
	}
	t.Fatal("No posts found")
	return nil
}

func getLastPostID(t *testing.T) string {
	postsMutex.RLock()
	defer postsMutex.RUnlock()

	var lastID string
	for id := range posts {
		lastID = id
	}
	return lastID
}

func isPostsSorted(posts []*model.Post) bool {
	return sort.SliceIsSorted(posts, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, posts[i].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, posts[j].CreatedAt)

		// Handle invalid dates by putting them at the end
		if err1 != nil || err2 != nil {
			return err1 == nil
		}
		return t1.After(t2)
	})
}

func addTestPosts(ctx context.Context, t *testing.T) {
	postsMutex.Lock()
	defer postsMutex.Unlock()

	for i := 0; i < 5; i++ {
		post := &model.Post{
			ID:        uuid.NewString(),
			Title:     fmt.Sprintf("Test Post %d", i),
			Content:   "Test Content",
			Author:    "Tester",
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
		}
		posts[post.ID] = post
	}
}
