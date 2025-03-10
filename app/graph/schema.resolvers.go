package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.66

import (
	"context"
	"fmt"
	"graphql-comment-system/app/graph/model"
	"graphql-comment-system/app/pkg/validator"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Post - resolver для поля post типа Comment.
// Отвечает за получение поста, к которому относится комментарий.
func (r *commentResolver) Post(ctx context.Context, obj *model.Comment) (*model.Post, error) {
	post, err := r.Resolver.PostStore.GetPostByID(ctx, obj.PostID)
	if err != nil {
		// В случае ошибки получения поста возвращаем ошибку с указанием ID поста.
		return nil, fmt.Errorf("post with id %s not found: %w", obj.PostID, err)
	}
	return post, nil
}

// Replies - resolver для поля replies типа Comment.
// Обеспечивает получение ответов на комментарий с пагинацией.
func (r *commentResolver) Replies(ctx context.Context, obj *model.Comment, first *int32, after *string) (*model.CommentConnection, error) {
	var n = int32(10) // Значение по умолчанию для количества возвращаемых ответов.
	if first != nil {
		n = int32(*first) // Если в запросе указано количество 'first', используем его.
	}
	result, err := r.Resolver.CommentStore.GetRepliesForComment(ctx, obj.ID, n, after)
	if err != nil {
		return nil, err // Возвращаем ошибку, если не удалось получить ответы.
	}

	var edges []*model.CommentEdge
	for _, edge := range result.Edges {
		edges = append(edges, &model.CommentEdge{
			Cursor: edge.Cursor,
			Node:   edge.Node,
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: result.PageInfo.HasNextPage,
		StartCursor: func() *string { // Функция для получения стартового курсора.
			if len(edges) > 0 {
				return &edges[0].Cursor // Возвращаем курсор первого элемента, если есть edges.
			}
			return nil // Иначе возвращаем nil.
		}(),
		EndCursor: func() *string { // Функция для получения конечного курсора.
			if len(edges) > 0 {
				return &edges[len(edges)-1].Cursor // Возвращаем курсор последнего элемента, если есть edges.
			}
			return nil // Иначе возвращаем nil.
		}(),
	}

	return &model.CommentConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

// CreatePost - resolver для мутации createPost.
// Создает новый пост в системе, предварительно валидируя входные данные.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	// Валидация входных данных для создания поста.
	validationErrors := validator.ValidateCreatePostInput(ctx, input.Title, input.Author, input.Content)
	if len(validationErrors) > 0 {
		var errorMessages []string
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		// Возвращаем ошибку, если валидация не пройдена, с перечислением ошибок.
		return nil, fmt.Errorf("validation errors: %s", strings.Join(errorMessages, "; "))
	}

	post := &model.Post{
		ID:            uuid.NewString(), // Генерация уникального ID для поста.
		Author:        input.Author,
		Title:         input.Title,
		Content:       input.Content,
		CreatedAt:     time.Now().Format(time.RFC3339), // Установка времени создания поста.
		AllowComments: input.AllowComments,
	}
	err := r.Resolver.PostStore.AddPost(ctx, post)
	if err != nil {
		// Возвращаем ошибку, если не удалось создать пост в хранилище.
		return nil, fmt.Errorf("error creating post: %w", err)
	}
	return post, nil // Возвращаем созданный пост.
}

// CreateComment - resolver для мутации createComment.
// Создает новый комментарий, включая валидацию входных данных и проверок связей (пост, родительский комментарий).
func (r *mutationResolver) CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error) {
	// Валидация входных данных для создания комментария.
	validationErrors := validator.ValidateCreateCommentInput(r.PostStore, r.CommentStore, ctx, input.Author, input.Content, input.PostID, input.ParentID)
	if len(validationErrors) > 0 {
		var errorMessages []string
		for _, err := range validationErrors {
			errorMessages = append(errorMessages, err.Error())
		}
		// Возвращаем ошибку, если валидация не пройдена.
		return nil, fmt.Errorf("validation errors: %s", strings.Join(errorMessages, "; "))
	}

	comment := &model.Comment{
		ID:        uuid.NewString(), // Генерация уникального ID для комментария.
		PostID:    input.PostID,
		Author:    input.Author,
		Content:   input.Content,
		CreatedAt: time.Now().Format(time.RFC3339), // Установка времени создания комментария.
		ParentID:  input.ParentID,
	}
	err := r.Resolver.CommentStore.AddComment(ctx, comment)
	if err != nil {
		// Возвращаем ошибку, если не удалось создать комментарий.
		return nil, fmt.Errorf("error creating comment: %w", err)
	}
	return comment, nil // Возвращаем созданный комментарий.
}

// Comments - resolver для поля comments типа Post.
// Позволяет получить комментарии к посту с поддержкой пагинации.
func (r *postResolver) Comments(ctx context.Context, obj *model.Post, first *int32, after *string) (*model.CommentConnection, error) {
	var n = int32(10) // Значение по умолчанию для количества возвращаемых комментариев.
	if first != nil {
		n = int32(*first) // Если в запросе указано 'first', используем его значение.
	}
	result, err := r.Resolver.CommentStore.GetCommentsForPost(ctx, obj.ID, n, after)
	if err != nil {
		return nil, err // Возвращаем ошибку, если не удалось получить комментарии.
	}

	var edges []*model.CommentEdge
	for _, edge := range result.Edges {
		edges = append(edges, &model.CommentEdge{
			Cursor: edge.Cursor,
			Node:   edge.Node,
		})
	}

	pageInfo := &model.PageInfo{
		HasNextPage: result.PageInfo.HasNextPage,
		StartCursor: func() *string { // Функция для определения стартового курсора.
			if len(edges) > 0 {
				return &edges[0].Cursor // Возвращаем курсор первого элемента, если есть edges.
			}
			return nil // Иначе возвращаем nil.
		}(),
		EndCursor: func() *string { // Функция для определения конечного курсора.
			if len(edges) > 0 {
				return &edges[len(edges)-1].Cursor // Возвращаем курсор последнего элемента, если есть edges.
			}
			return nil // Иначе возвращаем nil.
		}(),
	}

	return &model.CommentConnection{
		Edges:    edges,
		PageInfo: pageInfo,
	}, nil
}

// Post - resolver для query post.
// Возвращает один пост по его ID.
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	post, err := r.Resolver.PostStore.GetPostByID(ctx, id)
	if err != nil {
		// В случае ошибки получения поста, возвращаем ошибку.
		return nil, fmt.Errorf("get post by id: %w", err)
	}
	return post, nil // Возвращаем найденный пост.
}

// Posts - resolver для query posts.
// Возвращает список постов с поддержкой пагинации.
func (r *queryResolver) Posts(ctx context.Context, first *int32, after *string) (*model.PostConnection, error) {
	var firstValue int32 = 10 // Значение по умолчанию для количества возвращаемых постов.
	if first != nil {
		firstValue = int32(*first) // Если в запросе указано 'first', используем его значение.
	}

	result, err := r.Resolver.PostStore.GetPosts(ctx, firstValue, after)
	if err != nil {
		// Возвращаем ошибку, если не удалось получить посты.
		return nil, fmt.Errorf("get posts: %w", err)
	}

	postEdges := make([]*model.PostEdge, len(result.Edges))
	for i, edge := range result.Edges {
		postEdges[i] = &model.PostEdge{
			Cursor: edge.Cursor,
			Node:   edge.Node,
		}
	}

	var hasPreviousPage bool
	if after != nil {
		hasPreviousPage = true // Если 'after' курсор указан, значит есть предыдущая страница.
	} else {
		hasPreviousPage = false // Иначе, предыдущей страницы нет (начало списка).
	}

	var startCursor, endCursor *string
	if len(postEdges) > 0 {
		startCursor = &postEdges[0].Cursor              // Курсор первого поста в списке.
		endCursor = &postEdges[len(postEdges)-1].Cursor // Курсор последнего поста в списке.
	}

	pageInfo := &model.PageInfo{
		HasNextPage:     result.PageInfo.HasNextPage,
		HasPreviousPage: hasPreviousPage,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	return &model.PostConnection{
		Edges:    postEdges,
		PageInfo: pageInfo,
	}, nil
}

// Comment - resolver для query comment.
// Возвращает один комментарий по его ID.
func (r *queryResolver) Comment(ctx context.Context, id string) (*model.Comment, error) {
	comment, err := r.Resolver.CommentStore.GetCommentByID(ctx, id)
	if err != nil {
		// Возвращаем ошибку, если не удалось получить комментарий.
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return comment, nil // Возвращаем найденный комментарий.
}

// Comment returns CommentResolver implementation.
func (r *Resolver) Comment() CommentResolver { return &commentResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Post returns PostResolver implementation.
func (r *Resolver) Post() PostResolver { return &postResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type commentResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type postResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
