package storage

import "github.com/StonerF/posts/graph/model"

type Repository interface {
	CreatePost(authorID, title, content string, allowComments bool) (*model.Post, error)
	GetPosts(limit int, after *string) (*model.PostConnection, error)
	GetPostByID(id string) (*model.Post, error)
	CreateComment(authorID, postID string, content string) (*model.Comment, error)
	GetComments(postID string, limit int, after *string) (*model.CommentConnection, error)
	CreateReply(authorID, postID string, content string, parentID *string) (*model.Comment, error)
	GetRepliesByCommentID(commentID string, limit int, after *string) (*model.CommentConnection, error)
}
