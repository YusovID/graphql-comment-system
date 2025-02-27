package postgres

import (
	"context"
	"fmt"
	"graphql-comment-system/app/graph/model"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

// CommentStore struct - структура, реализующая хранилище комментариев.
type CommentStore struct {
	conn *pgx.Conn // conn - указатель на соединение с базой данных PostgreSQL.
}

// NewCommentStore - функция-конструктор для создания нового экземпляра CommentStore.
func NewCommentStore(conn *pgx.Conn) *CommentStore {
	return &CommentStore{
		conn: conn, // Инициализация хранилища с переданным соединением.
	}
}

// AddComment - метод для добавления нового комментария в хранилище.
func (c *CommentStore) AddComment(ctx context.Context, comment *model.Comment) error {
	// Выполнение SQL-запроса для вставки нового комментария в таблицу 'comments'.
	_, err := c.conn.Exec(ctx, `INSERT INTO comments (id, post_id, parent_id, author, content, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		comment.ID, comment.PostID, comment.ParentID, comment.Author, comment.Content, comment.CreatedAt)
	if err != nil {
		// В случае ошибки при вставке, возвращается ошибка с контекстом.
		return fmt.Errorf("error inserting comment: %w", err)
	}

	return nil // В случае успешного добавления комментария, возвращается nil.
}

// GetCommentByID - метод для получения комментария по его уникальному идентификатору.
func (c *CommentStore) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	// Выполнение SQL-запроса для выбора комментария из таблицы 'comments' по ID.
	row := c.conn.QueryRow(ctx, `SELECT id, post_id, parent_id, author, content, created_at FROM comments WHERE id = $1`, id)
	var comment model.Comment // Объявление переменной для хранения результата запроса.

	// Сканирование данных из строки результата запроса в структуру comment.
	err := row.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Content, &comment.CreatedAt)
	if err != nil {
		// В случае ошибки при получении комментария, возвращается nil и ошибка с контекстом.
		return nil, fmt.Errorf("error getting comment by ID: %w", err)
	}

	return &comment, nil // В случае успеха, возвращается указатель на найденный комментарий и nil.
}

// GetCommentsForPost - метод для получения комментариев к определенному посту с поддержкой пагинации.
func (c *CommentStore) GetCommentsForPost(ctx context.Context, postID string, first int32, after *string) (*model.CommentConnection, error) {
	// Выполнение SQL-запроса для выбора комментариев к посту из таблицы 'comments'.
	rows, err := c.conn.Query(ctx, `SELECT id, post_id, parent_id, author, content, created_at FROM comments WHERE post_id = $1`, postID)
	if err != nil {
		// В случае ошибки при запросе комментариев, возвращается nil и ошибка с контекстом.
		return nil, fmt.Errorf("error getting comments for post: %w", err)
	}

	defer rows.Close() // Ensure rows are closed after function execution.

	var comments []*model.Comment // Объявление слайса для хранения списка комментариев.

	// Итерация по каждой строке, возвращенной запросом.
	for rows.Next() {
		var comment model.Comment // Объявление переменной для хранения текущего комментария.

		// Сканирование данных из текущей строки в структуру comment.
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Content, &comment.CreatedAt)
		if err != nil {
			// В случае ошибки при сканировании комментария, возвращается nil и ошибка с контекстом.
			return nil, fmt.Errorf("error scanning comments: %w", err)
		}

		comments = append(comments, &comment) // Добавление сканированного комментария в список.
	}
	// Сортировка комментариев по дате создания, от старых к новым.
	sort.Slice(comments, func(i, j int) bool {
		timeI, err := time.Parse(time.RFC3339, comments[i].CreatedAt) // Парсинг времени создания первого комментария.
		if err != nil {
			return false // В случае ошибки парсинга, порядок не меняется.
		}

		timeJ, err := time.Parse(time.RFC3339, comments[j].CreatedAt) // Парсинг времени создания второго комментария.
		if err != nil {
			return true // В случае ошибки парсинга, порядок меняется.
		}

		return timeI.Before(timeJ) // Возвращает true, если время создания первого комментария раньше второго.
	})

	var commentEdges []*model.CommentEdge // Объявление слайса для хранения edges комментариев.

	// Преобразование списка комментариев в список edges для GraphQL Connection.
	for _, comment := range comments {
		commentEdges = append(commentEdges, &model.CommentEdge{
			Node:   comment,    // Установка комментария в качестве ноды edge.
			Cursor: comment.ID, // Установка ID комментария в качестве курсора edge.
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: false, // По умолчанию, предполагается отсутствие следующей страницы.
	}

	return &model.CommentConnection{
		Edges:    commentEdges, // Установка списка edges в Connection.
		PageInfo: pageInfo,     // Установка PageInfo в Connection.
	}, nil // Возвращается структура CommentConnection и nil в случае успеха.
}

// GetRepliesForComment - метод для получения ответов на конкретный комментарий (ветка ответов).
func (c *CommentStore) GetRepliesForComment(ctx context.Context, commentID string, first int32, after *string) (*model.CommentConnection, error) {
	// Выполнение SQL-запроса для выбора комментариев-ответов на комментарий из таблицы 'comments'.
	rows, err := c.conn.Query(ctx, `SELECT id, post_id, parent_id, author, content, created_at FROM comments WHERE parent_id = $1`, commentID)
	if err != nil {
		// В случае ошибки при запросе ответов, возвращается nil и ошибка с контекстом.
		return nil, fmt.Errorf("error getting replies for comment: %w", err)
	}

	defer rows.Close() // Ensure rows are closed after function execution.

	var comments []*model.Comment // Объявление слайса для хранения списка комментариев-ответов.

	// Итерация по каждой строке, возвращенной запросом.
	for rows.Next() {
		var comment model.Comment // Объявление переменной для хранения текущего комментария-ответа.

		// Сканирование данных из текущей строки в структуру comment.
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Content, &comment.CreatedAt)
		if err != nil {
			// В случае ошибки при сканировании комментария-ответа, возвращается nil и ошибка с контекстом.
			return nil, fmt.Errorf("error scanning comments: %w", err)
		}

		comments = append(comments, &comment) // Добавление сканированного комментария-ответа в список.
	}

	// Сортировка комментариев-ответов по дате создания, от старых к новым.
	sort.Slice(comments, func(i, j int) bool {
		timeI, err := time.Parse(time.RFC3339, comments[i].CreatedAt) // Парсинг времени создания первого комментария-ответа.
		if err != nil {
			return false // В случае ошибки парсинга, порядок не меняется.
		}

		timeJ, err := time.Parse(time.RFC3339, comments[j].CreatedAt) // Парсинг времени создания второго комментария-ответа.
		if err != nil {
			return true // В случае ошибки парсинга, порядок меняется.
		}

		return timeI.Before(timeJ) // Возвращает true, если время создания первого комментария-ответа раньше второго.
	})

	var commentEdges []*model.CommentEdge // Объявление слайса для хранения edges комментариев-ответов.

	// Преобразование списка комментариев-ответов в список edges для GraphQL Connection.
	for _, comment := range comments {
		commentEdges = append(commentEdges, &model.CommentEdge{
			Node:   comment,    // Установка комментария-ответа в качестве ноды edge.
			Cursor: comment.ID, // Установка ID комментария-ответа в качестве курсора edge.
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: false, // По умолчанию, предполагается отсутствие следующей страницы.
	}

	return &model.CommentConnection{
		Edges:    commentEdges, // Установка списка edges в Connection.
		PageInfo: pageInfo,     // Установка PageInfo в Connection.
	}, nil // Возвращается структура CommentConnection и nil в случае успеха.
}
