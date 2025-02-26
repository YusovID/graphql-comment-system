// pkg/validator/validator.go
package validator

import (
	"context"
	"fmt"
	inmemory "graphql-comment-system/pkg/data/in-memory"
	"strings"
)

// ValidationError представляет ошибку валидации.
type ValidationError struct {
	Field   string
	Message string
}

// Error реализует интерфейс error.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateCreatePostInput проверяет входные данные для создания поста.
func ValidateCreatePostInput(ctx context.Context, title, author, content string) []error {
	var errors []error

	// Проверка на пустые значения
	if len(strings.TrimSpace(title)) == 0 {
		errors = append(errors, &ValidationError{Field: "title", Message: "title cannot be empty"})
	}

	if len(strings.TrimSpace(author)) == 0 {
		errors = append(errors, &ValidationError{Field: "author", Message: "author cannot be empty"})
	}

	if len(strings.TrimSpace(content)) == 0 {
		errors = append(errors, &ValidationError{Field: "content", Message: "content cannot be empty"})
	}

	return errors
}

// ValidateCreateCommentInput проверяет входные данные для создания комментария.
func ValidateCreateCommentInput(ctx context.Context, author, content, postId string, parentId *string) []error {
	var errors []error

	// Проверка на пустые значения
	if len(strings.TrimSpace(author)) == 0 {
		errors = append(errors, &ValidationError{Field: "author", Message: "author cannot be empty"})
	}

	if len(strings.TrimSpace(content)) == 0 {
		errors = append(errors, &ValidationError{Field: "content", Message: "content cannot be empty"})
	}
	if len(content) > 2000 {
		errors = append(errors, &ValidationError{Field: "content", Message: "comment cannot be longer than 2000 characters"})
	}

	if len(strings.TrimSpace(postId)) == 0 {
		errors = append(errors, &ValidationError{Field: "postId", Message: "postId cannot be empty"})
	}

	// Проверка на существование поста
	_, err := inmemory.GetPostByID(ctx, postId)
	if err != nil {
		errors = append(errors, &ValidationError{Field: "postId", Message: "post with id " + postId + " not found"})
	}

	if parentId != nil {
		if len(strings.TrimSpace(*parentId)) == 0 {
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parentId cannot be empty"})
		}

		comment, err := inmemory.GetCommentByID(ctx, *parentId)
		if err != nil {
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parent comment with id " + *parentId + " not found"})
			return errors
		}

		if comment.PostID != postId {
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parent comment with id " + *parentId + " does not belong to post with id " + postId})
		}
	}

	return errors
}
