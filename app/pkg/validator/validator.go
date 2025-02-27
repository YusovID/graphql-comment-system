package validator

import (
	"context"
	"fmt"
	"graphql-comment-system/app/pkg/data"
	"strings"
)

// ValidationError - структура для представления ошибки валидации.
type ValidationError struct {
	Field   string // Field - название поля, в котором произошла ошибка.
	Message string // Message - сообщение об ошибке валидации.
}

// Error - реализация интерфейса error для ValidationError.
// Позволяет использовать ValidationError как тип ошибки.
func (e *ValidationError) Error() string {
	// Форматирование ошибки в виде строки, содержащей поле и сообщение об ошибке.
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateCreatePostInput - функция для валидации входных данных при создании поста.
// Проверяет обязательные поля: title, author, content на заполненность.
func ValidateCreatePostInput(ctx context.Context, title, author, content string) []error {
	var errors []error // errors - слайс для хранения ошибок валидации.

	// Проверка поля title на пустоту.
	if len(strings.TrimSpace(title)) == 0 {
		// Если title состоит только из пробелов или пустое, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "title", Message: "title cannot be empty"})
	}

	// Проверка поля author на пустоту.
	if len(strings.TrimSpace(author)) == 0 {
		// Если author состоит только из пробелов или пустое, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "author", Message: "author cannot be empty"})
	}

	// Проверка поля content на пустоту.
	if len(strings.TrimSpace(content)) == 0 {
		// Если content состоит только из пробелов или пустое, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "content", Message: "content cannot be empty"})
	}

	return errors // Возвращает слайс накопленных ошибок валидации.
}

// ValidateCreateCommentInput - функция для валидации входных данных при создании комментария.
// Выполняет несколько проверок: обязательные поля, максимальную длину контента,
// существование поста и родительского комментария (при наличии).
func ValidateCreateCommentInput(postStore data.PostStore, commentStore data.CommentStore, ctx context.Context, author, content, postId string, parentId *string) []error {
	var errors []error // errors - слайс для хранения ошибок валидации.

	// Проверка поля author на пустоту.
	if len(strings.TrimSpace(author)) == 0 {
		// Если author состоит только из пробелов или пустое, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "author", Message: "author cannot be empty"})
	}

	// Проверка поля content на пустоту.
	if len(strings.TrimSpace(content)) == 0 {
		// Если content состоит только из пробелов или пустое, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "content", Message: "content cannot be empty"})
	}
	// Проверка максимальной длины поля content.
	if len(content) > 2000 {
		// Если длина content превышает 2000 символов, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "content", Message: "comment cannot be longer than 2000 characters"})
	}

	// Проверка поля postId на пустоту.
	if len(strings.TrimSpace(postId)) == 0 {
		// Если postId состоит только из пробелов или пустое, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "postId", Message: "postId cannot be empty"})
	}

	// Проверка существования поста с указанным postId в хранилище.
	_, err := postStore.GetPostByID(ctx, postId)
	if err != nil {
		// Если пост с указанным postId не найден, добавляется ошибка валидации.
		errors = append(errors, &ValidationError{Field: "postId", Message: "post with id " + postId + " not found"})
	}

	// Проверка поля parentId, если оно передано (не nil).
	if parentId != nil {
		// Проверка parentId на пустоту, если указан.
		if len(strings.TrimSpace(*parentId)) == 0 {
			// Если parentId состоит только из пробелов или пустое, добавляется ошибка валидации.
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parentId cannot be empty"})
		}

		// Проверка существования родительского комментария с указанным parentId в хранилище.
		comment, err := commentStore.GetCommentByID(ctx, *parentId)
		if err != nil {
			// Если родительский комментарий с указанным parentId не найден, добавляется ошибка валидации и функция завершает работу.
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parent comment with id " + *parentId + " not found"})
			return errors // Возврат ошибок, так как дальнейшая проверка не имеет смысла без родительского комментария.
		}

		// Проверка принадлежности родительского комментария к тому же посту, что и текущий комментарий.
		if comment.PostID != postId {
			// Если PostID родительского комментария не совпадает с PostID текущего комментария, добавляется ошибка валидации.
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parent comment with id " + *parentId + " does not belong to post with id " + postId})
		}
	}

	return errors // Возвращает слайс накопленных ошибок валидации.
}