package data

import (
	"context"
	"graphql-comment-system/app/graph/model"
)

// PostStore определяет интерфейс для хранилища данных постов.
// Этот интерфейс абстрагирует способ доступа к данным постов, позволяя использовать различные реализации хранения данных.
type PostStore interface {
	// AddPost добавляет новый пост в хранилище данных.
	// Принимает контекст для управления временем выполнения и структуру Post для сохранения.
	// Возвращает ошибку, если добавление поста не удалось.
	AddPost(ctx context.Context, post *model.Post) error
	// GetPostByID извлекает пост из хранилища данных по его уникальному идентификатору.
	// Принимает контекст и строковый ID поста.
	// Возвращает структуру Post и ошибку, если пост не найден или произошла ошибка при извлечении.
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	// GetPosts извлекает список постов из хранилища данных с поддержкой пагинации.
	// Принимает контекст, количество постов для извлечения (`first`) и курсор (`after`) для пагинации.
	// Возвращает структуру PostConnection, содержащую список постов и информацию о пагинации, а также ошибку в случае ошибки.
	GetPosts(ctx context.Context, first int32, after *string) (*model.PostConnection, error)
}

// CommentStore определяет интерфейс для хранилища данных комментариев.
// Аналогично PostStore, этот интерфейс абстрагирует логику доступа к данным комментариев.
type CommentStore interface {
	// AddComment добавляет новый комментарий в хранилище данных.
	// Принимает контекст и структуру Comment для сохранения.
	// Возвращает ошибку, если добавление комментария не удалось.
	AddComment(ctx context.Context, comment *model.Comment) error
	// GetCommentByID извлекает комментарий из хранилища данных по его уникальному идентификатору.
	// Принимает контекст и строковый ID комментария.
	// Возвращает структуру Comment и ошибку, если комментарий не найден или произошла ошибка.
	GetCommentByID(ctx context.Context, id string) (*model.Comment, error)
	// GetCommentsForPost извлекает список комментариев для определенного поста с поддержкой пагинации.
	// Принимает контекст, ID поста, количество комментариев (`first`) и курсор (`after`).
	// Возвращает CommentConnection с комментариями и информацией о пагинации для указанного поста, и ошибку в случае неудачи.
	GetCommentsForPost(ctx context.Context, postID string, first int32, after *string) (*model.CommentConnection, error)
	// GetRepliesForComment извлекает список ответов (дочерних комментариев) для определенного комментария с пагинацией.
	// Принимает контекст, ID родительского комментария, количество ответов (`first`) и курсор (`after`).
	// Возвращает CommentConnection с ответами и информацией о пагинации для указанного комментария, и ошибку в случае ошибки.
	GetRepliesForComment(ctx context.Context, commentID string, first int32, after *string) (*model.CommentConnection, error)
}
