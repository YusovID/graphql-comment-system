// pkg/validator/validator.go
package validator

import (
	"context"
	"fmt"
	"graphql-comment-system/app/pkg/data"
	"strings"
)

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidateCreatePostInput validates the input data for creating a post.
func ValidateCreatePostInput(ctx context.Context, title, author, content string) []error {
	var errors []error

	// Check for empty values
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

// ValidateCreateCommentInput validates the input data for creating a comment.
func ValidateCreateCommentInput(postStore data.PostStore, commentStore data.CommentStore, ctx context.Context, author, content, postId string, parentId *string) []error {
	var errors []error

	// Check for empty values
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

	// Check for post existence
	_, err := postStore.GetPostByID(ctx, postId) // Use the interface method
	if err != nil {
		errors = append(errors, &ValidationError{Field: "postId", Message: "post with id " + postId + " not found"})
	}

	if parentId != nil {
		if len(strings.TrimSpace(*parentId)) == 0 {
			errors = append(errors, &ValidationError{Field: "parentId", Message: "parentId cannot be empty"})
		}

		comment, err := commentStore.GetCommentByID(ctx, *parentId) //Use the interface method
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
