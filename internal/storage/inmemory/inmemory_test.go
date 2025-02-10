package inmemory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePost(t *testing.T) {
	repo := NewInMemoryRepository()

	post, err := repo.CreatePost("1", "Title", "Content", true)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "1", post.AuthorID)
	assert.Equal(t, "Title", post.Title)
	assert.Equal(t, "Content", post.Content)
	assert.Equal(t, true, post.AllowComments)
	assert.NotNil(t, post.ID)
}

func TestGetPosts(t *testing.T) {
	repo := NewInMemoryRepository()

	repo.CreatePost("1", "Title1", "Content1", true)
	repo.CreatePost("2", "Title2", "Content2", false)

	conn, err := repo.GetPosts(10, nil)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 2)
	assert.Equal(t, "Title1", conn.Edges[0].Node.Title)
	assert.Equal(t, "Title2", conn.Edges[1].Node.Title)
}

func TestGetPostByID(t *testing.T) {
	repo := NewInMemoryRepository()

	post, _ := repo.CreatePost("1", "Title", "Content", true)

	fetchedPost, err := repo.GetPostByID(post.ID)

	assert.NoError(t, err)
	assert.NotNil(t, fetchedPost)
	assert.Equal(t, post.ID, fetchedPost.ID)
}

func TestCreateComment(t *testing.T) {
	repo := NewInMemoryRepository()

	post, _ := repo.CreatePost("1", "Title", "Content", true)

	comment, err := repo.CreateComment("2", post.ID, "Nice post!")

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "2", comment.AuthorID)
	assert.Equal(t, post.ID, comment.PostID)
	assert.Equal(t, "Nice post!", comment.Content)
	assert.NotEmpty(t, comment.ID)
	assert.NotEmpty(t, comment.CreatedAt)
}

func TestGetComments(t *testing.T) {
	repo := NewInMemoryRepository()

	post, _ := repo.CreatePost("1", "Title", "Content", true)
	repo.CreateComment("2", post.ID, "Nice post!")
	repo.CreateComment("3", post.ID, "I agree!")

	conn, err := repo.GetComments(post.ID, 10, nil)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 2)
	assert.Equal(t, "Nice post!", conn.Edges[0].Node.Content)
	assert.Equal(t, "I agree!", conn.Edges[1].Node.Content)
}

func TestCreateReply(t *testing.T) {
	repo := NewInMemoryRepository()

	post, _ := repo.CreatePost("1", "Title", "Content", true)
	comment, _ := repo.CreateComment("2", post.ID, "Nice post!")

	reply, err := repo.CreateReply("3", post.ID, "Thanks!", &comment.ID)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, "3", reply.AuthorID)
	assert.Equal(t, post.ID, reply.PostID)
	assert.Equal(t, comment.ID, *reply.ParentID)
	assert.Equal(t, "Thanks!", reply.Content)
	assert.NotEmpty(t, reply.ID)
	assert.NotEmpty(t, reply.CreatedAt)
}

func TestGetRepliesByCommentID(t *testing.T) {
	repo := NewInMemoryRepository()

	post, _ := repo.CreatePost("1", "Title", "Content", true)
	comment, _ := repo.CreateComment("2", post.ID, "Nice post!")
	repo.CreateReply("3", post.ID, "Thanks!", &comment.ID)

	conn, err := repo.GetRepliesByCommentID(comment.ID, 10, nil)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 1)
	assert.Equal(t, "Thanks!", conn.Edges[0].Node.Content)
}

func TestGetPostsWithPagination(t *testing.T) {
	repo := NewInMemoryRepository()

	repo.CreatePost("1", "Title1", "Content1", true)
	repo.CreatePost("2", "Title2", "Content2", false)
	repo.CreatePost("3", "Title3", "Content3", true)

	conn, err := repo.GetPosts(2, nil)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 2)
	assert.Equal(t, "Title1", conn.Edges[0].Node.Title)
	assert.Equal(t, "Title2", conn.Edges[1].Node.Title)

	after := conn.PageInfo.EndCursor
	conn, err = repo.GetPosts(2, after)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 1)
	assert.Equal(t, "Title3", conn.Edges[0].Node.Title)
}

func TestGetCommentsWithPagination(t *testing.T) {
	repo := NewInMemoryRepository()

	post, _ := repo.CreatePost("1", "Title", "Content", true)
	repo.CreateComment("2", post.ID, "Nice post!")
	repo.CreateComment("3", post.ID, "I agree!")
	repo.CreateComment("4", post.ID, "Thanks!")

	conn, err := repo.GetComments(post.ID, 2, nil)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 2)
	assert.Equal(t, "Nice post!", conn.Edges[0].Node.Content)
	assert.Equal(t, "I agree!", conn.Edges[1].Node.Content)

	after := conn.PageInfo.EndCursor
	conn, err = repo.GetComments(post.ID, 2, after)

	assert.NoError(t, err)
	assert.Len(t, conn.Edges, 1)
	assert.Equal(t, "Thanks!", conn.Edges[0].Node.Content)
}
