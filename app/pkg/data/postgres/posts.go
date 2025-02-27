package postgres

import (
	"context"
	"fmt"
	"graphql-comment-system/graph/model"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostStore struct {
	conn *pgx.Conn
}

func NewPostStore(conn *pgx.Conn) *PostStore {
	return &PostStore{
		conn: conn,
	}
}

func (p *PostStore) AddPost(ctx context.Context, post *model.Post) error {
	_, err := p.conn.Exec(ctx, `INSERT INTO posts (id, author, title, content, created_at, allow_comments) VALUES ($1, $2, $3, $4, $5, $6)`,
		post.ID, post.Author, post.Title, post.Content, post.CreatedAt, post.AllowComments)
	if err != nil {
		return fmt.Errorf("error inserting post: %w", err)
	}
	return nil
}

func (p *PostStore) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	row := p.conn.QueryRow(ctx, `SELECT id, author, title, content, created_at, allow_comments FROM posts WHERE id = $1`, id)
	var post model.Post
	err := row.Scan(&post.ID, &post.Author, &post.Title, &post.Content, &post.CreatedAt, &post.AllowComments)
	if err != nil {
		return nil, fmt.Errorf("error getting post by ID: %w", err)
	}
	return &post, nil
}

func (p *PostStore) GetPosts(ctx context.Context, first int32, after *string) (*model.PostConnection, error) {
	rows, err := p.conn.Query(ctx, `SELECT id, author, title, content, created_at, allow_comments FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("error getting posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Author, &post.Title, &post.Content, &post.CreatedAt, &post.AllowComments)
		if err != nil {
			return nil, fmt.Errorf("error scanning posts: %w", err)
		}
		posts = append(posts, &post)
	}
	sort.Slice(posts, func(i, j int) bool {
		timeI, err := time.Parse(time.RFC3339, posts[i].CreatedAt)
		if err != nil {
			return false
		}
		timeJ, err := time.Parse(time.RFC3339, posts[j].CreatedAt)
		if err != nil {
			return true
		}
		return timeI.Before(timeJ)
	})
	// Handle pagination (not fully implemented here)
	var postEdges []*model.PostEdge
	for _, post := range posts {
		postEdges = append(postEdges, &model.PostEdge{
			Node:   post,
			Cursor: post.ID,
		})
	}
	pageInfo := &model.PageInfo{
		HasNextPage: false,
	}

	return &model.PostConnection{
		Edges:    postEdges,
		PageInfo: pageInfo,
	}, nil
}
