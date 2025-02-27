package migration

import "time"

type Post struct {
	ID            string    `db:"id"`
	Author        string    `db:"author"`
	Title         string    `db:"title"`
	Content       string    `db:"content"`
	CreatedAt     time.Time `db:"created_at"`
	AllowComments bool      `db:"allow_comments"`
}

type Comment struct {
	ID        string    `db:"id"`
	PostID    string    `db:"post_id"`
	Post      *Post     `db:"-"`
	ParentID  *string   `db:"parent_id"`
	Author    string    `db:"author"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
