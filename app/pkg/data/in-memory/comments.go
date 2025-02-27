package inmemory

import (
	"context"
	"fmt"
	"graphql-comment-system/app/graph/model"
	"sort"
	"sync"
	"time"
)

// CommentStore реализует интерфейс data.CommentStore для хранения комментариев в памяти.
type CommentStore struct{}

// NewCommentStore создает и возвращает новый экземпляр CommentStore.
func NewCommentStore() *CommentStore {
	return &CommentStore{}
}

// comments хранит комментарии в памяти в виде map, где ключ - ID комментария.
var comments map[string]*model.Comment
// commentsMutex обеспечивает потокобезопасный доступ к map comments.
var commentsMutex sync.RWMutex

// CommentsResult структура для возврата списка комментариев с информацией о пагинации (не используется в текущей реализации).
type CommentsResult struct {
	Comments    []*model.Comment
	HasNextPage bool
}

// InitializeComments инициализирует хранилище in-memory тестовыми данными комментариев.
// Используется для демонстрации и локальной разработки.
func InitializeComments() {
	comments = make(map[string]*model.Comment)
	ctx := context.Background()
	store := NewCommentStore()

	// Создание тестового комментария 1
	comment1 := &model.Comment{
		ID:        "1",
		Author:    "Комментатор 1",
		Content:   "Отличный первый пост!",
		CreatedAt: time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
		PostID:    "1",
	}
	store.AddComment(ctx, comment1)

	// Создание тестового комментария 2
	comment2 := &model.Comment{
		ID:        "2",
		Author:    "Комментатор 2",
		Content:   "Согласен, очень интересно!",
		CreatedAt: time.Now().Add(-time.Hour).Format(time.RFC3339),
		PostID:    "1",
	}
	store.AddComment(ctx, comment2)

	// Создание тестового комментария 3
	comment3 := &model.Comment{
		ID:        "3",
		Author:    "Комментатор 3",
		Content:   "Мне не очень...",
		CreatedAt: time.Now().Format(time.RFC3339),
		PostID:    "2",
	}
	store.AddComment(ctx, comment3)
}

// GetCommentByID возвращает комментарий из in-memory хранилища по его ID.
// Возвращает ошибку, если комментарий не найден.
func (*CommentStore) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	commentsMutex.RLock() // Блокировка на чтение для обеспечения конкурентного доступа.
	defer commentsMutex.RUnlock()

	comment, ok := comments[id]
	if !ok {
		return nil, fmt.Errorf("comment with id %s not found", id) // Комментарий с указанным ID не найден.
	}

	return comment, nil
}

// GetCommentsForPost возвращает список комментариев для указанного поста с поддержкой пагинации.
// `first` - количество комментариев для возврата, `afterCursor` - курсор для пагинации.
func (*CommentStore) GetCommentsForPost(ctx context.Context, postID string, first int32, afterCursor *string) (*model.CommentConnection, error) {
	commentsMutex.RLock() // Блокировка на чтение для обеспечения конкурентного доступа.
	defer commentsMutex.RUnlock()

	filtered := make([]*model.Comment, 0)

	// Фильтрация комментариев по PostID
	for _, comment := range comments {
		if comment.PostID == postID {
			filtered = append(filtered, comment)
		}
	}

	// Сортировка комментариев по дате создания (CreatedAt) в порядке возрастания.
	sort.Slice(filtered, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, filtered[i].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, filtered[j].CreatedAt)
		if err1 != nil || err2 != nil {
			return false // В случае ошибки парсинга дат, порядок не меняем.
		}

		return t1.Before(t2) // Сортировка по возрастанию CreatedAt.
	})

	// Проверка корректности дат (лишняя операция, можно убрать в production, если уверены в данных).
	for _, c := range filtered {
		_, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing date: %w", err) // Ошибка парсинга даты.
		}
	}

	start := 0
	// Если указан курсор `afterCursor`, находим индекс, с которого начинать выборку.
	if afterCursor != nil {
		for i, c := range filtered {
			if c.ID == *afterCursor {
				start = i + 1 // Начинаем с элемента, следующего за курсором.
				break
			}
		}
	}

	end := start + int(first) // Вычисление конечного индекса для нарезки слайса.
	if end > len(filtered) {
		end = len(filtered) // Если `end` выходит за границы слайса, обрезаем до конца.
	}

	commentSlice := filtered[start:end] // Нарезка слайса комментариев для текущей страницы.

	hasNextPage := end < len(filtered) // Определение, есть ли следующая страница.

	return &model.CommentConnection{
		Edges:     convertToCommentEdges(commentSlice), // Преобразование слайса комментариев в слайс edges.
		PageInfo: &model.PageInfo{HasNextPage: hasNextPage}, // Информация о пагинации.
	}, nil
}

// convertToCommentEdges преобразует слайс комментариев в слайс CommentEdge для GraphQL Connection.
func convertToCommentEdges(comments []*model.Comment) []*model.CommentEdge {
	edges := make([]*model.CommentEdge, len(comments))

	// Для каждого комментария создается CommentEdge, содержащий курсор и ноду (комментарий).
	for i, comment := range comments {
		edges[i] = &model.CommentEdge{
			Node: comment, // Установка ноды (комментария).
		}
	}

	return edges
}

// AddComment добавляет новый комментарий в in-memory хранилище.
func (*CommentStore) AddComment(ctx context.Context, comment *model.Comment) error {
	commentsMutex.Lock() // Блокировка на запись для обеспечения эксклюзивного доступа к map.
	defer commentsMutex.Unlock()

	comments[comment.ID] = comment // Добавление комментария в map comments.

	return nil
}

// GetRepliesForComment возвращает ответы на комментарий (рекурсивные комментарии) с пагинацией.
// `parentID` - ID родительского комментария, `first` - количество ответов, `afterCursor` - курсор.
func (*CommentStore) GetRepliesForComment(ctx context.Context, parentID string, first int32, afterCursor *string) (*model.CommentConnection, error) {
	commentsMutex.RLock() // Блокировка на чтение для обеспечения конкурентного доступа.
	defer commentsMutex.RUnlock()

	filtered := make([]*model.Comment, 0)

	// Фильтрация комментариев для получения ответов на конкретный родительский комментарий.
	for _, comment := range comments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			filtered = append(filtered, comment)
		}
	}

	// Сортировка ответов по дате создания (CreatedAt) в порядке возрастания.
	sort.Slice(filtered, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, filtered[i].CreatedAt)
		t2, err2 := time.Parse(time.RFC3339, filtered[j].CreatedAt)
		if err1 != nil || err2 != nil {
			return false // В случае ошибки парсинга дат, порядок не меняем.
		}

		return t1.Before(t2) // Сортировка по возрастанию CreatedAt.
	})

	// Проверка корректности дат (лишняя операция, можно убрать в production).
	for _, c := range filtered {
		_, err := time.Parse(time.RFC3339, c.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error parsing date: %w", err) // Ошибка парсинга даты.
		}
	}

	start := 0
	// Обработка курсора `afterCursor` для пагинации.
	if afterCursor != nil {
		for i, c := range filtered {
			if c.ID == *afterCursor {
				start = i + 1 // Начинаем с элемента, следующего за курсором.
				break
			}
		}
	}

	end := start + int(first) // Вычисление конечного индекса.
	if end > len(filtered) {
		end = len(filtered) // Обрезка до конца слайса, если `end` выходит за границы.
	}

	commentSlice := filtered[start:end] // Нарезка слайса ответов.

	hasNextPage := end < len(filtered) // Определение, есть ли следующая страница ответов.

	return &model.CommentConnection{
		Edges:     convertToCommentEdges(commentSlice), // Преобразование в CommentEdge.
		PageInfo: &model.PageInfo{HasNextPage: hasNextPage}, // Информация о пагинации.
	}, nil
}