package inmemory

import (
	"context"
	"graphql-comment-system/app/graph/model"
	"testing"
	"time"
)

func TestNewCommentStore(t *testing.T) {
	store := NewCommentStore()
	if store == nil {
		t.Error("Expected non-nil CommentStore instance")
	}
}

func setupTestEnvironment() {
	comments = make(map[string]*model.Comment)
}

func TestInitializeComments(t *testing.T) {
	InitializeComments()

	if len(comments) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(comments))
	}

	// Проверка существования комментариев
	comment1, exists := comments["1"]
	if !exists {
		t.Error("Comment with ID 1 not found")
	}

	if comment1.Author != "Комментатор 1" || comment1.PostID != "1" {
		t.Errorf("Comment 1 has incorrect data: %+v", comment1)
	}

	comment3, exists := comments["3"]
	if !exists {
		t.Error("Comment with ID 3 not found")
	}

	if comment3.Author != "Комментатор 3" || comment3.PostID != "2" {
		t.Errorf("Comment 3 has incorrect data: %+v", comment3)
	}
}

func TestGetCommentByID(t *testing.T) {
	setupTestEnvironment()
	store := NewCommentStore()
	ctx := context.Background()

	testComment := &model.Comment{
		ID:        "42",
		Author:    "Тестовый автор",
		Content:   "Тестовый контент",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "5",
	}

	err := store.AddComment(ctx, testComment)
	if err != nil {
		t.Fatalf("Failed to add test comment: %v", err)
	}

	// Тест получения существующего комментария
	comment, err := store.GetCommentByID(ctx, "42")
	if err != nil {
		t.Errorf("Failed to get comment by ID: %v", err)
	}

	if comment.ID != "42" || comment.Author != "Тестовый автор" {
		t.Errorf("Retrieved comment doesn't match expected: %+v", comment)
	}

	// Тест получения несуществующего комментария
	_, err = store.GetCommentByID(ctx, "999")
	if err == nil {
		t.Error("Expected error when getting non-existent comment, got nil")
	}
}

func TestAddComment(t *testing.T) {
	setupTestEnvironment()
	store := NewCommentStore()
	ctx := context.Background()

	testComment := &model.Comment{
		ID:        "123",
		Author:    "Тестовый автор",
		Content:   "Тестовый контент",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "5",
	}

	err := store.AddComment(ctx, testComment)
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err)
	}

	// Проверяем, что комментарий добавлен
	commentsMutex.RLock()
	comment, exists := comments["123"]
	commentsMutex.RUnlock()

	if !exists {
		t.Error("Comment was not added to the store")
	}

	if comment.Author != "Тестовый автор" || comment.Content != "Тестовый контент" {
		t.Errorf("Added comment has incorrect data: %+v", comment)
	}
}

func TestGetCommentsForPost(t *testing.T) {
	setupTestEnvironment()
	store := NewCommentStore()
	ctx := context.Background()

	// Добавляем несколько комментариев для одного поста
	for i := 0; i < 5; i++ {
		comment := &model.Comment{
			ID:        string(rune('A' + i)),
			Author:    "Автор " + string(rune('A'+i)),
			Content:   "Контент " + string(rune('A'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			PostID:    "post1",
		}
		store.AddComment(ctx, comment)
	}

	// Добавляем комментарий для другого поста
	otherComment := &model.Comment{
		ID:        "X",
		Author:    "Другой автор",
		Content:   "Другой контент",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "post2",
	}
	store.AddComment(ctx, otherComment)

	// Тест получения всех комментариев для поста
	connection, err := store.GetCommentsForPost(ctx, "post1", 10, nil)
	if err != nil {
		t.Fatalf("Failed to get comments for post: %v", err)
	}

	if len(connection.Edges) != 5 {
		t.Errorf("Expected 5 comments for post1, got %d", len(connection.Edges))
	}

	// Проверка пагинации
	connection, err = store.GetCommentsForPost(ctx, "post1", 2, nil)
	if err != nil {
		t.Fatalf("Failed to get paginated comments: %v", err)
	}

	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 comments for pagination, got %d", len(connection.Edges))
	}

	if !connection.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true for paginated results")
	}

	// Проверка курсора
	firstCommentID := connection.Edges[0].Node.ID
	connection, err = store.GetCommentsForPost(ctx, "post1", 2, &firstCommentID)
	if err != nil {
		t.Fatalf("Failed to get comments after cursor: %v", err)
	}

	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 comments after cursor, got %d", len(connection.Edges))
	}

	if connection.Edges[0].Node.ID == firstCommentID {
		t.Error("Expected comments after cursor to exclude the cursor comment")
	}
}

func TestGetRepliesForComment(t *testing.T) {
	setupTestEnvironment()
	store := NewCommentStore()
	ctx := context.Background()

	// Добавляем родительский комментарий
	parentComment := &model.Comment{
		ID:        "parent1",
		Author:    "Родитель",
		Content:   "Родительский комментарий",
		CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
		PostID:    "post1",
	}
	store.AddComment(ctx, parentComment)

	// Добавляем ответы на комментарий
	for i := 0; i < 4; i++ {
		parentID := "parent1"
		reply := &model.Comment{
			ID:        "reply" + string(rune('1'+i)),
			Author:    "Ответ " + string(rune('1'+i)),
			Content:   "Содержание ответа " + string(rune('1'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
			PostID:    "post1",
			ParentID:  &parentID,
		}
		store.AddComment(ctx, reply)
	}

	// Добавляем ответ на другой комментарий
	otherParentID := "parent2"
	otherReply := &model.Comment{
		ID:        "otherReply",
		Author:    "Другой ответ",
		Content:   "Ответ на другой комментарий",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "post1",
		ParentID:  &otherParentID,
	}
	store.AddComment(ctx, otherReply)

	// Тест получения всех ответов на комментарий
	connection, err := store.GetRepliesForComment(ctx, "parent1", 10, nil)
	if err != nil {
		t.Fatalf("Failed to get replies for comment: %v", err)
	}

	if len(connection.Edges) != 4 {
		t.Errorf("Expected 4 replies for parent1, got %d", len(connection.Edges))
	}

	// Проверка пагинации
	connection, err = store.GetRepliesForComment(ctx, "parent1", 2, nil)
	if err != nil {
		t.Fatalf("Failed to get paginated replies: %v", err)
	}

	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 replies for pagination, got %d", len(connection.Edges))
	}

	if !connection.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true for paginated replies")
	}

	// Проверка курсора
	firstReplyID := connection.Edges[0].Node.ID
	connection, err = store.GetRepliesForComment(ctx, "parent1", 2, &firstReplyID)
	if err != nil {
		t.Fatalf("Failed to get replies after cursor: %v", err)
	}

	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 replies after cursor, got %d", len(connection.Edges))
	}
}
