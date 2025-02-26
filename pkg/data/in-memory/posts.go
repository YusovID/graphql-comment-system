package inmemory

import (
	"context"
	"fmt"
	"graphql-comment-system/graph/model"
	"sort"
	"strconv"
	"sync"
	"time"
)

type PostsResult struct {
	Posts       []*model.Post
	HasNextPage bool
}

var posts map[string]*model.Post
var postsMutex sync.RWMutex
//TODO: Удалить
var postID int = 2

func InitializePosts() {
	posts = make(map[string]*model.Post)
	ctx := context.Background()

	post1 := &model.Post{
		ID:            "1",
		Title:         "Первый пост",
		Content:       "Содержание первого поста",
		Author:        "Автор 1",
		CreatedAt:     time.Now().Add(time.Hour).Format(time.RFC3339),
		AllowComments: true,
	}
	AddPost(ctx, post1)

	post2 := &model.Post{
		ID:            "2",
		Title:         "Второй пост",
		Content:       "Содержание второго поста",
		Author:        "Автор 2",
		CreatedAt:     time.Now().Format(time.RFC3339),
		AllowComments: false,
	}
	AddPost(ctx, post2)
}

func GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	postsMutex.RLock()
	defer postsMutex.RUnlock()

	post, ok := posts[id]
	if !ok {
		return nil, fmt.Errorf("post with id %s not found", id)
	}
	return post, nil
}

func GetPosts(ctx context.Context, first int32, afterCursor *string) (PostsResult, error) {
	postsMutex.RLock()
	defer postsMutex.RUnlock()

	var validPosts []*model.Post

	for _, post := range posts {
		_, err := time.Parse(time.RFC3339, post.CreatedAt)
		if err != nil {
			// Возвращаем ошибку, если не удалось распарсить дату
			return PostsResult{}, fmt.Errorf("invalid CreatedAt format for post %s: %w", post.ID, err)
		}
		validPosts = append(validPosts, post)
	}

	sort.SliceStable(validPosts, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, validPosts[i].CreatedAt)
		t2, _ := time.Parse(time.RFC3339, validPosts[j].CreatedAt)
		return t1.After(t2)
	})

	sortedPosts := validPosts // Теперь используем только валидные посты

	startIndex := 0
	if afterCursor != nil && *afterCursor != "" {
		afterID := *afterCursor

		for index, post := range sortedPosts {
			if post.ID == afterID {
				startIndex = index + 1
				break
			}
		}
	}

	numPosts := int32(len(sortedPosts) - startIndex)
	if first == 0 || first > numPosts {
		first = numPosts
	}

	postSlice := sortedPosts[startIndex : startIndex+int(first)]
	hasNextPage := numPosts > first

	return PostsResult{
		Posts:       postSlice,
		HasNextPage: hasNextPage,
	}, nil
}

func AddPost(ctx context.Context, post *model.Post) {
	postsMutex.Lock()
	defer postsMutex.Unlock()
	posts[post.ID] = post
}

//TODO: Удалить
func GetNextPostID() string {
	ID := strconv.Itoa(postID)
	commentID++
	return ID
}
