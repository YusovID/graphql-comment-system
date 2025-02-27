package inmemory

import (
	"context"
	"graphql-comment-system/graph/model"
	"testing"
	"time"
)

func TestNewPostStore(t *testing.T) {
	store := NewPostStore()
	if store == nil {
		t.Error("Expected non-nil PostStore instance")
	}
}

func setupPostTestEnvironment() {
	posts = make(map[string]*model.Post)
}

func TestInitializePosts(t *testing.T) {
	InitializePosts()
	
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts))
	}
	
	// Проверка существования постов
	post1, exists := posts["1"]
	if !exists {
		t.Error("Post with ID 1 not found")
	}
	
	if post1.Author != "Автор 1" || post1.Title != "Первый пост" || !post1.AllowComments {
		t.Errorf("Post 1 has incorrect data: %+v", post1)
	}
	
	post2, exists := posts["2"]
	if !exists {
		t.Error("Post with ID 2 not found")
	}
	
	if post2.Author != "Автор 2" || post2.Title != "Второй пост" || post2.AllowComments {
		t.Errorf("Post 2 has incorrect data: %+v", post2)
	}
}

func TestGetPostByID(t *testing.T) {
	setupPostTestEnvironment()
	store := NewPostStore()
	ctx := context.Background()
	
	testPost := &model.Post{
		ID:            "42",
		Title:         "Тестовый пост",
		Content:       "Содержание тестового поста",
		Author:        "Тестовый автор",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: true,
	}
	
	err := store.AddPost(ctx, testPost)
	if err != nil {
		t.Fatalf("Failed to add test post: %v", err)
	}
	
	// Тест получения существующего поста
	post, err := store.GetPostByID(ctx, "42")
	if err != nil {
		t.Errorf("Failed to get post by ID: %v", err)
	}
	
	if post.ID != "42" || post.Title != "Тестовый пост" || post.Author != "Тестовый автор" {
		t.Errorf("Retrieved post doesn't match expected: %+v", post)
	}
	
	// Тест получения несуществующего поста
	_, err = store.GetPostByID(ctx, "999")
	if err == nil {
		t.Error("Expected error when getting non-existent post, got nil")
	}
}

func TestAddPost(t *testing.T) {
	setupPostTestEnvironment()
	store := NewPostStore()
	ctx := context.Background()
	
	testPost := &model.Post{
		ID:            "123",
		Title:         "Тестовый пост",
		Content:       "Содержание тестового поста",
		Author:        "Тестовый автор",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: true,
	}
	
	err := store.AddPost(ctx, testPost)
	if err != nil {
		t.Fatalf("Failed to add post: %v", err)
	}
	
	// Проверяем, что пост добавлен
	postsMutex.RLock()
	post, exists := posts["123"]
	postsMutex.RUnlock()
	
	if !exists {
		t.Error("Post was not added to the store")
	}
	
	if post.Title != "Тестовый пост" || post.Author != "Тестовый автор" {
		t.Errorf("Added post has incorrect data: %+v", post)
	}
}

func TestGetPosts(t *testing.T) {
	setupPostTestEnvironment()
	store := NewPostStore()
	ctx := context.Background()
	
	// Добавляем несколько постов с разными датами создания
	for i := 0; i < 5; i++ {
		post := &model.Post{
			ID:            string(rune('A' + i)),
			Title:         "Заголовок " + string(rune('A'+i)),
			Content:       "Содержание " + string(rune('A'+i)),
			Author:        "Автор " + string(rune('A'+i)),
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			AllowComments: i%2 == 0,
		}
		store.AddPost(ctx, post)
	}
	
	// Тест получения всех постов
	connection, err := store.GetPosts(ctx, 10, nil)
	if err != nil {
		t.Fatalf("Failed to get posts: %v", err)
	}
	
	if len(connection.Edges) != 5 {
		t.Errorf("Expected 5 posts, got %d", len(connection.Edges))
	}
	
	// Проверка сортировки по дате (сначала новые)
	for i := 0; i < len(connection.Edges)-1; i++ {
		time1, _ := time.Parse(time.RFC3339, connection.Edges[i].Node.CreatedAt)
		time2, _ := time.Parse(time.RFC3339, connection.Edges[i+1].Node.CreatedAt)
		if !time1.After(time2) {
			t.Error("Posts are not properly sorted by creation date (newest first)")
		}
	}
	
	// Проверка пагинации
	connection, err = store.GetPosts(ctx, 2, nil)
	if err != nil {
		t.Fatalf("Failed to get paginated posts: %v", err)
	}
	
	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 posts for pagination, got %d", len(connection.Edges))
	}
	
	if !connection.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true for paginated results")
	}
	
	// Проверка курсора
	firstPostID := connection.Edges[0].Node.ID
	connection, err = store.GetPosts(ctx, 2, &firstPostID)
	if err != nil {
		t.Fatalf("Failed to get posts after cursor: %v", err)
	}
	
	if len(connection.Edges) == 0 {
		t.Error("Expected posts after cursor, got none")
	} else if connection.Edges[0].Node.ID == firstPostID {
		t.Error("Expected posts after cursor to exclude the cursor post")
	}
	
	// Тест с некорректным форматом даты
	badPost := &model.Post{
		ID:            "BadDate",
		Title:         "Пост с некорректной датой",
		Content:       "Содержание",
		Author:        "Автор",
		CreatedAt:     "неправильная дата",
		AllowComments: true,
	}
	
	err = store.AddPost(ctx, badPost)
	if err != nil {
		t.Fatalf("Failed to add post with bad date: %v", err)
	}
	
	_, err = store.GetPosts(ctx, 10, nil)
	if err == nil {
		t.Error("Expected error for post with invalid date format, got nil")
	}
}