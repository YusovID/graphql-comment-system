package inmemory

import (
	"context"
	"fmt"
	"graphql-comment-system/app/graph/model"
	"sort"
	"sync"
	"time"
)

// PostStore реализует интерфейс data.PostStore для хранения постов в памяти.
// Используется как временное хранилище данных или для целей разработки и тестирования.
type PostStore struct{}

// NewPostStore создает и возвращает новый экземпляр PostStore.
// Функция является конструктором для структуры PostStore.
func NewPostStore() *PostStore {
	return &PostStore{}
}

// PostsResult структура для хранения результатов запроса списка постов.
// Включает в себя слайс постов и флаг, указывающий на наличие следующей страницы данных.
type PostsResult struct {
	Posts       []*model.Post
	HasNextPage bool
}

// posts хранит посты в памяти в виде map, где ключ - ID поста, а значение - указатель на структуру Post.
var posts map[string]*model.Post
// postsMutex используется для обеспечения потокобезопасности при работе с map posts.
// Это необходимо, так как к хранилищу могут обращаться несколько горутин одновременно.
var postsMutex sync.RWMutex

// InitializePosts инициализирует in-memory хранилище тестовыми данными для постов.
// Данные используются для демонстрационных целей и при отсутствии постоянного хранилища данных.
func InitializePosts() {
	posts = make(map[string]*model.Post)
	ctx := context.Background()
	store := NewPostStore()

	// Создание и добавление первого тестового поста.
	post1 := &model.Post{
		ID:            "1",
		Title:         "Первый пост",
		Content:       "Содержание первого поста",
		Author:        "Автор 1",
		CreatedAt:     time.Now().Add(time.Hour).Format(time.RFC3339),
		AllowComments: true, // Разрешены комментарии к посту.
	}
	store.AddPost(ctx, post1)

	// Создание и добавление второго тестового поста.
	post2 := &model.Post{
		ID:            "2",
		Title:         "Второй пост",
		Content:       "Содержание второго поста",
		Author:        "Автор 2",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: false, // Комментарии к посту запрещены.
	}
	store.AddPost(ctx, post2)
}

// GetPostByID извлекает пост из in-memory хранилища по его уникальному идентификатору.
// Возвращает пост и ошибку, если пост с указанным ID не найден.
func (*PostStore) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	postsMutex.RLock() // Устанавливаем блокировку на чтение, так как только читаем данные.
	defer postsMutex.RUnlock()

	post, ok := posts[id] // Пытаемся получить пост из map по его ID.
	if !ok {
		// Если пост не найден в map, возвращаем ошибку.
		return nil, fmt.Errorf("post with id %s not found", id)
	}

	return post, nil // Возвращаем найденный пост.
}

// GetPosts получает список постов из in-memory хранилища с поддержкой пагинации.
// `first` указывает количество постов, которые необходимо вернуть.
// `afterCursor` используется для пагинации, указывая, с какого элемента начать выборку.
func (*PostStore) GetPosts(ctx context.Context, first int32, afterCursor *string) (*model.PostConnection, error) {
	postsMutex.RLock() // Устанавливаем блокировку на чтение для обеспечения конкурентного доступа.
	defer postsMutex.RUnlock()

	var validPosts []*model.Post // Слайс для хранения валидных постов (с корректной датой).

	// Итерируем по всем постам в map.
	for _, post := range posts {
		// Проверяем формат даты CreatedAt для каждого поста.
		_, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			// Если формат даты некорректный, возвращаем ошибку с указанием ID поста и описанием проблемы.
			return nil, fmt.Errorf("invalid CreatedAt format for post %s: %w", post.ID, err)
		}

		validPosts = append(validPosts, post) // Добавляем пост в слайс валидных постов.
	}

	// Сортируем слайс валидных постов по дате создания в обратном порядке (сначала новые).
	sort.SliceStable(validPosts, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, validPosts[i].CreatedAt)
		t2, _ := time.Parse(time.RFC3339, validPosts[j].CreatedAt)
		return t1.After(t2) // Сортировка: более поздняя дата идет раньше.
	})

	sortedPosts := validPosts // Используем слайс отсортированных валидных постов.

	startIndex := 0 // Начальный индекс для нарезки слайса (для пагинации).

	// Если указан курсор `afterCursor`, определяем начальный индекс для следующей страницы.
	if afterCursor != nil && *afterCursor != "" {
		afterID := *afterCursor // ID поста, после которого нужно начать выборку.

		// Ищем индекс поста, ID которого соответствует `afterCursor`.
		for index, post := range sortedPosts {
			if post.ID == afterID {
				startIndex = index + 1 // Начинаем со следующего элемента после найденного.
				break
			}
		}
	}

	numPosts := int32(len(sortedPosts) - startIndex) // Количество постов, доступных для выборки, начиная с `startIndex`.

	// Если запрошенное количество `first` равно 0 или превышает доступное количество,
	// устанавливаем `first` равным количеству доступных постов.
	if first == 0 || first > numPosts {
		first = numPosts
	}

	postSlice := sortedPosts[startIndex : startIndex+int(first)] // Нарезаем слайс постов для текущей страницы.
	hasNextPage := numPosts > first                             // Определяем, есть ли следующая страница данных.

	// Возвращаем структуру PostConnection, содержащую edges (границы, в нашем случае - посты) и PageInfo (информация о пагинации).
	return &model.PostConnection{
		Edges: func() []*model.PostEdge { // Анонимная функция для создания слайса PostEdge.
			var edges []*model.PostEdge

			// Преобразуем слайс постов в слайс PostEdge.
			for _, post := range postSlice {
				edges = append(edges, &model.PostEdge{Node: post}) // Создаем PostEdge для каждого поста.
			}

			return edges
		}(),
		PageInfo: &model.PageInfo{
			HasNextPage: hasNextPage, // Устанавливаем флаг HasNextPage в PageInfo.
		},
	}, nil
}

// AddPost добавляет новый пост в in-memory хранилище.
func (*PostStore) AddPost(ctx context.Context, post *model.Post) error {
	postsMutex.Lock() // Устанавливаем блокировку на запись, так как изменяем map posts.
	defer postsMutex.Unlock()

	posts[post.ID] = post // Добавляем пост в map posts, используя ID поста как ключ.

	return nil
}