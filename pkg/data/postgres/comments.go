package postgres

import (
	"context"
	"fmt"
	"graphql-comment-system/graph/model"
	"sort"
	"time"
)

func (c *CommentStore) AddComment(ctx context.Context, comment *model.Comment) error {
	_, err := c.conn.Exec(ctx, `INSERT INTO comments (id, post_id, parent_id, author, content, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		comment.ID, comment.PostID, comment.ParentID, comment.Author, comment.Content, comment.CreatedAt)
	if err != nil {
		return fmt.Errorf("error inserting comment: %w", err)
	}
	return nil
}

func (c *CommentStore) GetCommentByID(ctx context.Context, id string) (*model.Comment, error) {
	row := c.conn.QueryRow(ctx, `SELECT id, post_id, parent_id, author, content, created_at FROM comments WHERE id = $1`, id)
	var comment model.Comment
	err := row.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("error getting comment by ID: %w", err)
	}
	return &comment, nil
}

func (c *CommentStore) GetCommentsForPost(ctx context.Context, postID string, first int32, after *string) (*model.CommentConnection, error) {
	rows, err := c.conn.Query(ctx, `SELECT id, post_id, parent_id, author, content, created_at FROM comments WHERE post_id = $1`, postID)
	if err != nil {
		return nil, fmt.Errorf("error getting comments for post: %w", err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning comments: %w", err)
		}
		comments = append(comments, &comment)
	}
	sort.Slice(comments, func(i, j int) bool {
		timeI, err := time.Parse(time.RFC3339, comments[i].CreatedAt)
		if err != nil {
			return false
		}
		timeJ, err := time.Parse(time.RFC3339, comments[j].CreatedAt)
		if err != nil {
			return true
		}
		return timeI.Before(timeJ)
	})
	// Handle pagination (not fully implemented here)
	var commentEdges []*model.CommentEdge
	for _, comment := range comments {
		commentEdges = append(commentEdges, &model.CommentEdge{
			Node:   comment,
			Cursor: comment.ID,
		})
	}
	pageInfo := &model.PageInfo{
		HasNextPage: false,
	}

	return &model.CommentConnection{
		Edges:    commentEdges,
		PageInfo: pageInfo,
	}, nil
}

func (c *CommentStore) GetRepliesForComment(ctx context.Context, commentID string, first int32, after *string) (*model.CommentConnection, error) {
	rows, err := c.conn.Query(ctx, `SELECT id, post_id, parent_id, author, content, created_at FROM comments WHERE parent_id = $1`, commentID)
	if err != nil {
		return nil, fmt.Errorf("error getting replies for comment: %w", err)
	}
	defer rows.Close()

	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.Author, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning comments: %w", err)
		}
		comments = append(comments, &comment)
	}
	sort.Slice(comments, func(i, j int) bool {
		timeI, err := time.Parse(time.RFC3339, comments[i].CreatedAt)
		if err != nil {
			return false
		}
		timeJ, err := time.Parse(time.RFC3339, comments[j].CreatedAt)
		if err != nil {
			return true
		}
		return timeI.Before(timeJ)
	})
	var commentEdges []*model.CommentEdge
	for _, comment := range comments {
		commentEdges = append(commentEdges, &model.CommentEdge{
			Node:   comment,
			Cursor: comment.ID,
		})
	}
	pageInfo := &model.PageInfo{
		HasNextPage: false,
	}

	return &model.CommentConnection{
		Edges:    commentEdges,
		PageInfo: pageInfo,
	}, nil
}
