package inmemory

import (
	"context"
	"graphql-comment-system/app/graph/model"
	"testing"
	"time"
)

func TestNewCommentStore(t *testing.T) {
	// Проверка создания нового экземпляра CommentStore.
	store := NewCommentStore()
	if store == nil {
		t.Error("Expected non-nil CommentStore instance") // Ожидался не nil экземпляр CommentStore.
	}
}

func setupTestEnvironment() {
	// Функция для настройки тестового окружения перед каждым тестом.
	// В данном случае, она инициализирует мапу comments для изоляции тестов.
	comments = make(map[string]*model.Comment)
}

func TestInitializeComments(t *testing.T) {
	// Тест функции InitializeComments, проверяющей начальное заполнение хранилища комментариями.
	InitializeComments()

	// Проверка, что после инициализации в хранилище должно быть ровно 3 комментария.
	if len(comments) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(comments)) // Ожидалось 3 комментария, получено %d.
	}

	// Проверка существования и корректности данных первого комментария (ID "1").
	comment1, exists := comments["1"]
	if !exists {
		t.Error("Comment with ID 1 not found") // Комментарий с ID 1 не найден.
	}

	// Проверка полей Author и PostID для комментария с ID "1".
	if comment1.Author != "Комментатор 1" || comment1.PostID != "1" {
		t.Errorf("Comment 1 has incorrect data: %+v", comment1) // Комментарий 1 содержит некорректные данные: %+v.
	}

	// Проверка существования и корректности данных третьего комментария (ID "3").
	comment3, exists := comments["3"]
	if !exists {
		t.Error("Comment with ID 3 not found") // Комментарий с ID 3 не найден.
	}

	// Проверка полей Author и PostID для комментария с ID "3".
	if comment3.Author != "Комментатор 3" || comment3.PostID != "2" {
		t.Errorf("Comment 3 has incorrect data: %+v", comment3) // Комментарий 3 содержит некорректные данные: %+v.
	}
}

func TestGetCommentByID(t *testing.T) {
	setupTestEnvironment() // Настройка тестового окружения.
	store := NewCommentStore() // Создание нового хранилища комментариев.
	ctx := context.Background() // Создание контекста для тестов.

	// Создание тестового комментария для добавления в хранилище.
	testComment := &model.Comment{
		ID:        "42",
		Author:    "Тестовый автор",
		Content:   "Тестовый контент",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "5",
	}

	// Добавление тестового комментария в хранилище.
	err := store.AddComment(ctx, testComment)
	if err != nil {
		t.Fatalf("Failed to add test comment: %v", err) // Не удалось добавить тестовый комментарий: %v.
	}

	// Тест получения существующего комментария по ID "42".
	comment, err := store.GetCommentByID(ctx, "42")
	if err != nil {
		t.Errorf("Failed to get comment by ID: %v", err) // Не удалось получить комментарий по ID: %v.
	}

	// Проверка, что полученный комментарий соответствует ожидаемому.
	if comment.ID != "42" || comment.Author != "Тестовый автор" {
		t.Errorf("Retrieved comment doesn't match expected: %+v", comment) // Полученный комментарий не соответствует ожидаемому: %+v.
	}

	// Тест получения несуществующего комментария (ID "999").
	_, err = store.GetCommentByID(ctx, "999")
	// Ожидается ошибка при попытке получить несуществующий комментарий.
	if err == nil {
		t.Error("Expected error when getting non-existent comment, got nil") // Ожидалась ошибка при получении несуществующего комментария, получен nil.
	}
}

func TestAddComment(t *testing.T) {
	setupTestEnvironment() // Настройка тестового окружения.
	store := NewCommentStore() // Создание нового хранилища комментариев.
	ctx := context.Background() // Создание контекста для тестов.

	// Создание тестового комментария для добавления.
	testComment := &model.Comment{
		ID:        "123",
		Author:    "Тестовый автор",
		Content:   "Тестовый контент",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "5",
	}

	// Добавление комментария в хранилище.
	err := store.AddComment(ctx, testComment)
	if err != nil {
		t.Fatalf("Failed to add comment: %v", err) // Не удалось добавить комментарий: %v.
	}

	// Проверка, что комментарий был успешно добавлен в хранилище.
	commentsMutex.RLock()
	comment, exists := comments["123"]
	commentsMutex.RUnlock()

	// Проверка, что комментарий с ID "123" существует в мапе comments.
	if !exists {
		t.Error("Comment was not added to the store") // Комментарий не был добавлен в хранилище.
	}

	// Проверка полей Author и Content добавленного комментария.
	if comment.Author != "Тестовый автор" || comment.Content != "Тестовый контент" {
		t.Errorf("Added comment has incorrect data: %+v", comment) // Добавленный комментарий содержит некорректные данные: %+v.
	}
}

func TestGetCommentsForPost(t *testing.T) {
	setupTestEnvironment() // Настройка тестового окружения.
	store := NewCommentStore() // Создание нового хранилища комментариев.
	ctx := context.Background() // Создание контекста для тестов.

	// Добавление нескольких комментариев для поста "post1".
	for i := 0; i < 5; i++ {
		comment := &model.Comment{
			ID:        string(rune('A' + i)), // ID комментария генерируется на основе буквы.
			Author:    "Автор " + string(rune('A'+i)),
			Content:   "Контент " + string(rune('A'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour).Format(time.RFC3339), // Время создания с разницей в час.
			PostID:    "post1", // Все комментарии для поста "post1".
		}
		store.AddComment(ctx, comment) // Добавление комментария в хранилище.
	}

	// Добавление комментария для другого поста "post2".
	otherComment := &model.Comment{
		ID:        "X",
		Author:    "Другой автор",
		Content:   "Другой контент",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "post2", // Комментарий для поста "post2".
	}
	store.AddComment(ctx, otherComment) // Добавление комментария в хранилище.

	// Тест получения всех комментариев для поста "post1".
	connection, err := store.GetCommentsForPost(ctx, "post1", 10, nil)
	if err != nil {
		t.Fatalf("Failed to get comments for post: %v", err) // Не удалось получить комментарии для поста: %v.
	}

	// Проверка, что количество полученных комментариев для "post1" равно 5.
	if len(connection.Edges) != 5 {
		t.Errorf("Expected 5 comments for post1, got %d", len(connection.Edges)) // Ожидалось 5 комментариев для post1, получено %d.
	}

	// Тест пагинации: запрашиваем первые 2 комментария для "post1".
	connection, err = store.GetCommentsForPost(ctx, "post1", 2, nil)
	if err != nil {
		t.Fatalf("Failed to get paginated comments: %v", err) // Не удалось получить пагинированные комментарии: %v.
	}

	// Проверка, что при пагинации получено 2 комментария.
	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 comments for pagination, got %d", len(connection.Edges)) // Ожидалось 2 комментария для пагинации, получено %d.
	}

	// Проверка наличия флага HasNextPage для пагинированных результатов.
	if !connection.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true for paginated results") // Ожидалось, что HasNextPage будет true для пагинированных результатов.
	}

	// Тест получения комментариев после курсора.
	firstCommentID := connection.Edges[0].Node.ID // ID первого комментария из предыдущего запроса.
	connection, err = store.GetCommentsForPost(ctx, "post1", 2, &firstCommentID) // Запрос комментариев после курсора.
	if err != nil {
		t.Fatalf("Failed to get comments after cursor: %v", err) // Не удалось получить комментарии после курсора: %v.
	}

	// Проверка, что получено 2 комментария после курсора.
	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 comments after cursor, got %d", len(connection.Edges)) // Ожидалось 2 комментария после курсора, получено %d.
	}

	// Проверка, что первый комментарий в новом запросе не совпадает с курсором (исключается сам курсор).
	if connection.Edges[0].Node.ID == firstCommentID {
		t.Error("Expected comments after cursor to exclude the cursor comment") // Ожидалось, что комментарии после курсора исключают комментарий-курсор.
	}
}

func TestGetRepliesForComment(t *testing.T) {
	setupTestEnvironment() // Настройка тестового окружения.
	store := NewCommentStore() // Создание нового хранилища комментариев.
	ctx := context.Background() // Создание контекста для тестов.

	// Создание родительского комментария.
	parentComment := &model.Comment{
		ID:        "parent1",
		Author:    "Родитель",
		Content:   "Родительский комментарий",
		CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
		PostID:    "post1",
	}
	store.AddComment(ctx, parentComment) // Добавление родительского комментария.

	// Добавление ответов на родительский комментарий "parent1".
	for i := 0; i < 4; i++ {
		parentID := "parent1" // Установка ParentID для ответов.
		reply := &model.Comment{
			ID:        "reply" + string(rune('1'+i)), // ID ответа генерируется на основе номера.
			Author:    "Ответ " + string(rune('1'+i)),
			Content:   "Содержание ответа " + string(rune('1'+i)),
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute).Format(time.RFC3339), // Время создания с разницей в минуты.
			PostID:    "post1", // Все ответы к посту "post1".
			ParentID:  &parentID, // Установка ParentID.
		}
		store.AddComment(ctx, reply) // Добавление ответа.
	}

	// Добавление ответа на другой родительский комментарий "parent2" (не влияет на тест).
	otherParentID := "parent2"
	otherReply := &model.Comment{
		ID:        "otherReply",
		Author:    "Другой ответ",
		Content:   "Ответ на другой комментарий",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "post1",
		ParentID:  &otherParentID, // ParentID = "parent2".
	}
	store.AddComment(ctx, otherReply) // Добавление ответа.

	// Тест получения всех ответов для комментария "parent1".
	connection, err := store.GetRepliesForComment(ctx, "parent1", 10, nil)
	if err != nil {
		t.Fatalf("Failed to get replies for comment: %v", err) // Не удалось получить ответы на комментарий: %v.
	}

	// Проверка, что количество полученных ответов для "parent1" равно 4.
	if len(connection.Edges) != 4 {
		t.Errorf("Expected 4 replies for parent1, got %d", len(connection.Edges)) // Ожидалось 4 ответа для parent1, получено %d.
	}

	// Тест пагинации ответов для "parent1": запрашиваем первые 2 ответа.
	connection, err = store.GetRepliesForComment(ctx, "parent1", 2, nil)
	if err != nil {
		t.Fatalf("Failed to get paginated replies: %v", err) // Не удалось получить пагинированные ответы: %v.
	}

	// Проверка, что при пагинации получено 2 ответа.
	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 replies for pagination, got %d", len(connection.Edges)) // Ожидалось 2 ответа для пагинации, получено %d.
	}

	// Проверка наличия флага HasNextPage для пагинированных ответов.
	if !connection.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true for paginated replies") // Ожидалось, что HasNextPage будет true для пагинированных ответов.
	}

	// Проверка курсора для ответов.
	firstReplyID := connection.Edges[0].Node.ID // ID первого ответа из предыдущего запроса.
	connection, err = store.GetRepliesForComment(ctx, "parent1", 2, &firstReplyID) // Запрос ответов после курсора.
	if err != nil {
		t.Fatalf("Failed to get replies after cursor: %v", err) // Не удалось получить ответы после курсора: %v.
	}

	// Проверка, что получено 2 ответа после курсора.
	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 replies after cursor, got %d", len(connection.Edges)) // Ожидалось 2 ответа после курсора, получено %d.
	}
}