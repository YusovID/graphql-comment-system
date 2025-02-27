package inmemory

import (
	"context"
	"graphql-comment-system/app/graph/model"
	"testing"
	"time"
)

func TestNewPostStore(t *testing.T) {
	// TestNewPostStore проверяет создание нового экземпляра PostStore.
	store := NewPostStore()
	if store == nil {
		t.Error("Expected non-nil PostStore instance") // Ошибка, если store == nil, ожидается не nil экземпляр.
	}
}

func setupPostTestEnvironment() {
	// setupPostTestEnvironment подготавливает тестовое окружение для тестов post store.
	// В данном случае, инициализирует мапу posts для изоляции тестов.
	posts = make(map[string]*model.Post)
}

func TestInitializePosts(t *testing.T) {
	// TestInitializePosts проверяет функцию InitializePosts, которая инициализирует тестовые посты.
	InitializePosts()

	// Проверяем, что после инициализации в мапе posts должно быть 2 поста.
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts, got %d", len(posts)) // Ожидалось 2 поста, получено %d.
	}

	// Проверка существования поста с ID "1".
	post1, exists := posts["1"]
	if !exists {
		t.Error("Post with ID 1 not found") // Ошибка, если пост с ID "1" не найден.
	}

	// Проверка корректности данных поста с ID "1".
	if post1.Author != "Автор 1" || post1.Title != "Первый пост" || !post1.AllowComments {
		t.Errorf("Post 1 has incorrect data: %+v", post1) // Ошибка, если данные поста "1" некорректны.
	}

	// Проверка существования поста с ID "2".
	post2, exists := posts["2"]
	if !exists {
		t.Error("Post with ID 2 not found") // Ошибка, если пост с ID "2" не найден.
	}

	// Проверка корректности данных поста с ID "2".
	if post2.Author != "Автор 2" || post2.Title != "Второй пост" || post2.AllowComments {
		t.Errorf("Post 2 has incorrect data: %+v", post2) // Ошибка, если данные поста "2" некорректны.
	}
}

func TestGetPostByID(t *testing.T) {
	setupPostTestEnvironment() // Настройка тестового окружения для постов.
	store := NewPostStore()     // Создание нового хранилища постов.
	ctx := context.Background() // Создание фонового контекста.

	// Создание тестового поста.
	testPost := &model.Post{
		ID:            "42",
		Title:         "Тестовый пост",
		Content:       "Содержание тестового поста",
		Author:        "Тестовый автор",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: true,
	}

	// Добавление тестового поста в хранилище.
	err := store.AddPost(ctx, testPost)
	if err != nil {
		t.Fatalf("Failed to add test post: %v", err) // Ошибка при добавлении тестового поста.
	}

	// Тест: получение существующего поста по ID "42".
	post, err := store.GetPostByID(ctx, "42")
	if err != nil {
		t.Errorf("Failed to get post by ID: %v", err) // Ошибка при получении поста по ID.
	}

	// Проверка, что полученный пост соответствует тестовому посту.
	if post.ID != "42" || post.Title != "Тестовый пост" || post.Author != "Тестовый автор" {
		t.Errorf("Retrieved post doesn't match expected: %+v", post) // Полученный пост не соответствует ожидаемому.
	}

	// Тест: получение несуществующего поста по ID "999".
	_, err = store.GetPostByID(ctx, "999")
	if err == nil {
		t.Error("Expected error when getting non-existent post, got nil") // Ожидалась ошибка при получении несуществующего поста, но ошибки нет.
	}
}

func TestAddPost(t *testing.T) {
	setupPostTestEnvironment() // Настройка тестового окружения для постов.
	store := NewPostStore()     // Создание нового хранилища постов.
	ctx := context.Background() // Создание фонового контекста.

	// Создание тестового поста.
	testPost := &model.Post{
		ID:            "123",
		Title:         "Тестовый пост",
		Content:       "Содержание тестового поста",
		Author:        "Тестовый автор",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: true,
	}

	// Добавление тестового поста в хранилище.
	err := store.AddPost(ctx, testPost)
	if err != nil {
		t.Fatalf("Failed to add post: %v", err) // Ошибка при добавлении поста.
	}

	// Проверка, что пост был успешно добавлен в хранилище.
	postsMutex.RLock()
	post, exists := posts["123"]
	postsMutex.RUnlock()

	// Проверяем, что пост с ID "123" существует.
	if !exists {
		t.Error("Post was not added to the store") // Ошибка, если пост не был добавлен в хранилище.
	}

	// Проверка корректности данных добавленного поста.
	if post.Title != "Тестовый пост" || post.Author != "Тестовый автор" {
		t.Errorf("Added post has incorrect data: %+v", post) // Ошибка, если данные добавленного поста некорректны.
	}
}

func TestGetPosts(t *testing.T) {
	setupPostTestEnvironment() // Настройка тестового окружения для постов.
	store := NewPostStore()     // Создание нового хранилища постов.
	ctx := context.Background() // Создание фонового контекста.

	// Добавление нескольких тестовых постов для тестирования пагинации и сортировки.
	for i := 0; i < 5; i++ {
		post := &model.Post{
			ID:            string(rune('A' + i)), // ID поста генерируется на основе буквы.
			Title:         "Заголовок " + string(rune('A'+i)),
			Content:       "Содержание " + string(rune('A'+i)),
			Author:        "Автор " + string(rune('A'+i)),
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour).Format(time.RFC3339), // Дата создания поста увеличивается на час с каждым постом.
			AllowComments: i%2 == 0,                                                                // Четные посты разрешают комментарии, нечетные - нет.
		}
		store.AddPost(ctx, post) // Добавление поста в хранилище.
	}

	// Тест: получение всех постов без пагинации.
	connection, err := store.GetPosts(ctx, 10, nil)
	if err != nil {
		t.Fatalf("Failed to get posts: %v", err) // Ошибка при получении постов.
	}

	// Проверка, что получено 5 постов.
	if len(connection.Edges) != 5 {
		t.Errorf("Expected 5 posts, got %d", len(connection.Edges)) // Ошибка, если количество полученных постов не равно 5.
	}

	// Проверка сортировки по дате создания (сначала самые новые).
	for i := 0; i < len(connection.Edges)-1; i++ {
		time1, _ := time.Parse(time.RFC3339, connection.Edges[i].Node.CreatedAt)
		time2, _ := time.Parse(time.RFC3339, connection.Edges[i+1].Node.CreatedAt)
		if !time1.After(time2) {
			t.Error("Posts are not properly sorted by creation date (newest first)") // Ошибка, если посты не отсортированы по дате создания (сначала новые).
		}
	}

	// Тест: пагинация - получение первых 2 постов.
	connection, err = store.GetPosts(ctx, 2, nil)
	if err != nil {
		t.Fatalf("Failed to get paginated posts: %v", err) // Ошибка при получении пагинированных постов.
	}

	// Проверка, что получено 2 поста при пагинации.
	if len(connection.Edges) != 2 {
		t.Errorf("Expected 2 posts for pagination, got %d", len(connection.Edges)) // Ошибка, если количество полученных постов при пагинации не равно 2.
	}

	// Проверка наличия флага HasNextPage для пагинированных результатов.
	if !connection.PageInfo.HasNextPage {
		t.Error("Expected HasNextPage to be true for paginated results") // Ошибка, если HasNextPage не true для пагинированных результатов.
	}

	// Тест: получение постов после курсора.
	firstPostID := connection.Edges[0].Node.ID // ID первого поста из предыдущего запроса для использования в качестве курсора.
	connection, err = store.GetPosts(ctx, 2, &firstPostID)
	if err != nil {
		t.Fatalf("Failed to get posts after cursor: %v", err) // Ошибка при получении постов после курсора.
	}

	// Проверка, что получены посты после курсора.
	if len(connection.Edges) == 0 {
		t.Error("Expected posts after cursor, got none") // Ошибка, если не получено постов после курсора.
	} else if connection.Edges[0].Node.ID == firstPostID {
		t.Error("Expected posts after cursor to exclude the cursor post") // Ошибка, если первый пост после курсора совпадает с курсором.
	}

	// Тест: обработка поста с некорректным форматом даты.
	badPost := &model.Post{
		ID:            "BadDate",
		Title:         "Пост с некорректной датой",
		Content:       "Содержание",
		Author:        "Автор",
		CreatedAt:     "неправильная дата", // Некорректный формат даты.
		AllowComments: true,
	}

	err = store.AddPost(ctx, badPost)
	if err != nil {
		t.Fatalf("Failed to add post with bad date: %v", err) // Ошибка при добавлении поста с некорректной датой.
	}

	// Ожидаем ошибку при получении постов, так как есть пост с некорректной датой.
	_, err = store.GetPosts(ctx, 10, nil)
	if err == nil {
		t.Error("Expected error for post with invalid date format, got nil") // Ошибка, если не получена ошибка при наличии поста с некорректной датой.
	}
}