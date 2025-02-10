package resolver_test

import (
	"context"
	"testing"
	"time"

	"github.com/StonerF/posts/internal/model"
	"github.com/StonerF/posts/internal/resolver"
	"github.com/StonerF/posts/internal/resolver/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostgresCreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockRepository(ctrl)

	authorID := "author123"
	title := "Test Post"
	content := "Lorem ipsum"
	allowComments := true

	mockDB.EXPECT().CreatePost(authorID, title, content, allowComments).Return(&model.Post{ID: "1", AuthorID: authorID, Title: title, Content: content, AllowComments: allowComments}, nil)

	res := resolver.NewResolver(mockDB)

	mures := res.Mutation()

	_, err := mures.CreatePost(context.Background(), authorID, title, content, allowComments)

	assert.NoError(t, err, "Expected no error")
}

func TestPostgresGetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockRepository(ctrl)

	postID := "post123"

	mockDB.EXPECT().GetPostByID(postID).Return(nil, nil)

	res := resolver.NewResolver(mockDB)

	mures := res.Query()

	_, err := mures.Post(context.Background(), postID)

	assert.NoError(t, err, "Expected no error")
}

func TestPostgresCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock.NewMockRepository(ctrl)

	authorID := "author123"
	postID := "post123"
	content := "This is a comment"

	mockDB.EXPECT().CreateComment(authorID, postID, content).Return(&model.Comment{ID: "1", AuthorID: authorID, PostID: postID, Content: content, CreatedAt: time.Now().String(), Replies: nil}, nil)

	mockDB.EXPECT().GetPostByID(postID).Return(&model.Post{ID: "1", AuthorID: authorID, Title: "Lorem", Content: "Lorem ipsum", AllowComments: true}, nil)

	res := resolver.NewResolver(mockDB)

	mures := res.Mutation()

	_, err := mures.CreateComment(context.Background(), authorID, postID, content)

	assert.NoError(t, err, "Expected no error")
}
