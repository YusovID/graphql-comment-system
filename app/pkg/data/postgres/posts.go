package postgres

import (
	"context"
	"fmt"
	"graphql-comment-system/app/graph/model"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

// PostStore struct - структура хранилища постов.
type PostStore struct {
	conn *pgx.Conn // conn - это экземпляр подключения к базе данных PostgreSQL.
}

// NewPostStore - функция-конструктор, возвращает новый экземпляр PostStore.
func NewPostStore(conn *pgx.Conn) *PostStore {
	return &PostStore{
		conn: conn, // Инициализация PostStore с переданным подключением к БД.
	}
}

// AddPost - метод для добавления нового поста в хранилище данных.
func (p *PostStore) AddPost(ctx context.Context, post *model.Post) error {
	// SQL-запрос для вставки данных нового поста в таблицу "posts".
	_, err := p.conn.Exec(ctx, `INSERT INTO posts (id, author, title, content, created_at, allow_comments) VALUES ($1, $2, $3, $4, $5, $6)`,
		post.ID, post.Author, post.Title, post.Content, post.CreatedAt, post.AllowComments)
	if err != nil {
		// В случае ошибки при выполнении SQL-запроса, возвращаем ошибку с форматированием.
		return fmt.Errorf("error inserting post: %w", err)
	}

	return nil // В случае успешного добавления поста, возвращаем nil (отсутствие ошибки).
}

// GetPostByID - метод для получения поста из хранилища данных по его уникальному идентификатору.
func (p *PostStore) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	// SQL-запрос для выбора всех полей поста из таблицы "posts" по заданному ID.
	row := p.conn.QueryRow(ctx, `SELECT id, author, title, content, created_at, allow_comments FROM posts WHERE id = $1`, id)
	var post model.Post // Объявляем переменную для хранения данных поста.

	// Сканируем данные из первой строки результата SQL-запроса в структуру 'post'.
	err := row.Scan(&post.ID, &post.Author, &post.Title, &post.Content, &post.CreatedAt, &post.AllowComments)
	if err != nil {
		// В случае ошибки сканирования данных (например, пост не найден), возвращаем ошибку.
		return nil, fmt.Errorf("error getting post by ID: %w", err)
	}

	return &post, nil // Возвращаем указатель на структуру 'post' с полученными данными.
}

// GetPosts - метод для получения списка постов из хранилища данных с поддержкой пагинации.
func (p *PostStore) GetPosts(ctx context.Context, first int32, after *string) (*model.PostConnection, error) {
	// SQL-запрос для выбора всех полей из таблицы "posts".
	rows, err := p.conn.Query(ctx, `SELECT id, author, title, content, created_at, allow_comments FROM posts`)
	if err != nil {
		// В случае ошибки выполнения SQL-запроса, возвращаем ошибку.
		return nil, fmt.Errorf("error getting posts: %w", err)
	}

	defer rows.Close() // Ensure rows are closed after function completion.

	var posts []*model.Post // Объявляем слайс для хранения указателей на структуры 'Post'.

	// Итерируем по каждой строке, возвращенной SQL-запросом.
	for rows.Next() {
		var post model.Post // Объявляем структуру для сканирования данных каждой строки.

		// Сканируем данные из текущей строки в структуру 'post'.
		err := rows.Scan(&post.ID, &post.Author, &post.Title, &post.Content, &post.CreatedAt, &post.AllowComments)
		if err != nil {
			// В случае ошибки сканирования, возвращаем ошибку.
			return nil, fmt.Errorf("error scanning posts: %w", err)
		}

		posts = append(posts, &post) // Добавляем полученный пост в слайс 'posts'.
	}

	// Сортируем слайс постов по дате создания (CreatedAt) в порядке возрастания.
	sort.Slice(posts, func(i, j int) bool {
		timeI, err := time.Parse(time.RFC3339, posts[i].CreatedAt) // Преобразуем строку даты в time.Time для поста i.
		if err != nil {
			return false // В случае ошибки парсинга даты, не меняем порядок (оставляем как есть).
		}

		timeJ, err := time.Parse(time.RFC3339, posts[j].CreatedAt) // Преобразуем строку даты в time.Time для поста j.
		if err != nil {
			return true // В случае ошибки парсинга даты, меняем порядок (j < i).
		}

		return timeI.Before(timeJ) // Возвращаем true, если дата поста i раньше даты поста j (сортировка по возрастанию).
	})

	var postEdges []*model.PostEdge // Объявляем слайс для хранения GraphQL edges для постов.

	// Итерируем по слайсу постов для преобразования в формат GraphQL Connection.
	for _, post := range posts {
		postEdges = append(postEdges, &model.PostEdge{ // Создаем новый PostEdge для каждого поста.
			Node:   post,    // Устанавливаем текущий пост в качестве Node.
			Cursor: post.ID, // В качестве курсора используем ID поста.
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: false, // В текущей реализации пагинация не реализована, HasNextPage всегда false.
	}

	return &model.PostConnection{
		Edges:    postEdges, // Устанавливаем слайс edges в PostConnection.
		PageInfo: pageInfo,  // Устанавливаем PageInfo в PostConnection.
	}, nil
}
