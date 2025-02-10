package inmemory

import (
	"errors"

	"strconv"
	"sync"
	"time"

	"github.com/StonerF/posts/internal/model"
)

type InMemoryRepository struct {
	posts    map[string]*model.Post
	comments map[string][]*model.Comment
	mutex    sync.RWMutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		posts:    make(map[string]*model.Post),
		comments: make(map[string][]*model.Comment),
	}
}

func (r *InMemoryRepository) CreatePost(authorID, title, content string, allowComments bool) (*model.Post, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	post := &model.Post{
		ID:            strconv.Itoa(len(r.posts) + 1),
		AuthorID:      authorID,
		Title:         title,
		Content:       content,
		AllowComments: allowComments,
	}

	r.posts[post.ID] = post

	return post, nil
}

func (r *InMemoryRepository) GetPosts(limit int, after *string) (*model.PostConnection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	posts := r.postsToSlice()

	startIndex := 0
	if after != nil {
		startIndex = r.findIndex(*after, posts)
		if startIndex == -1 {
			return nil, errors.New("invalid cursor")
		}
		startIndex++
	}

	endIndex := len(posts)
	if limit > 0 && limit < len(posts)-startIndex {
		endIndex = startIndex + limit
	}

	var edges []*model.PostEdge
	for _, post := range posts[startIndex:endIndex] {
		edges = append(edges, &model.PostEdge{
			Cursor: post.ID,
			Node:   post,
		})
	}

	hasNextPage := endIndex < len(posts)

	return &model.PostConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   r.getEndCursor(posts[startIndex:endIndex]),
			HasNextPage: hasNextPage,
		},
	}, nil
}

func (r *InMemoryRepository) GetPostByID(id string) (*model.Post, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	post, ok := r.posts[id]
	if !ok {
		return nil, errors.New("post not found")
	}

	return post, nil
}

func (r *InMemoryRepository) CreateComment(authorID, postID string, content string) (*model.Comment, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, ok := r.posts[postID]
	if !ok {
		return nil, errors.New("post not found")
	}

	comment := &model.Comment{
		ID:        strconv.Itoa(len(r.comments[postID]) + 1),
		AuthorID:  authorID,
		PostID:    postID,
		Content:   content,
		CreatedAt: time.Now().Format(time.RFC3339),
		Replies:   &model.CommentConnection{Edges: []*model.CommentEdge{}},
	}

	r.comments[postID] = append(r.comments[postID], comment)

	return comment, nil
}

func (r *InMemoryRepository) GetComments(postID string, limit int, after *string) (*model.CommentConnection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	comments, ok := r.comments[postID]
	if !ok {
		return nil, errors.New("comments not found")
	}

	startIndex := 0
	if after != nil {
		startIndex = r.findCommentIndex(*after, comments)
		if startIndex == -1 {
			return nil, errors.New("invalid cursor")
		}
		startIndex++
	}

	endIndex := len(comments)
	if limit > 0 && limit < len(comments)-startIndex {
		endIndex = startIndex + limit
	}

	var edges []*model.CommentEdge
	for _, comment := range comments[startIndex:endIndex] {
		edges = append(edges, &model.CommentEdge{
			Cursor: comment.ID,
			Node:   comment,
		})
	}

	hasNextPage := endIndex < len(comments)

	return &model.CommentConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   r.getEndCursorComments(comments[startIndex:endIndex]),
			HasNextPage: hasNextPage,
		},
	}, nil
}

func (r *InMemoryRepository) CreateReply(authorID, postID string, content string, parentID *string) (*model.Comment, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	comments, ok := r.comments[postID]
	if !ok {
		return nil, errors.New("post not found")
	}

	var parent *model.Comment
	for _, c := range comments {
		if c.ID == *parentID {
			parent = c
			break
		}
	}

	if parent == nil {
		return nil, errors.New("parent comment not found")
	}

	reply := &model.Comment{
		ID:        strconv.Itoa(len(comments) + 1),
		AuthorID:  authorID,
		PostID:    postID,
		ParentID:  parentID,
		Content:   content,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if parent.Replies == nil {
		parent.Replies = &model.CommentConnection{Edges: []*model.CommentEdge{}}
	}

	parent.Replies.Edges = append(parent.Replies.Edges, &model.CommentEdge{
		Cursor: reply.ID,
		Node:   reply,
	})

	r.comments[postID] = append(comments, reply)

	return reply, nil
}

func (r *InMemoryRepository) GetRepliesByCommentID(commentID string, limit int, after *string) (*model.CommentConnection, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var parentComment *model.Comment
	for _, comments := range r.comments {
		for _, comment := range comments {
			if comment.ID == commentID {
				parentComment = comment
				break
			}
		}
		if parentComment != nil {
			break
		}
	}

	if parentComment == nil {
		return nil, errors.New("comment not found")
	}

	replies := parentComment.Replies
	if replies == nil {
		return &model.CommentConnection{
			Edges:    []*model.CommentEdge{},
			PageInfo: &model.PageInfo{HasNextPage: false},
		}, nil
	}

	startIndex := 0
	if after != nil {
		startIndex = r.findReplyIndex(*after, replies.Edges)
		if startIndex == -1 {
			return nil, errors.New("invalid cursor")
		}
		startIndex++
	}

	endIndex := len(replies.Edges)
	if limit > 0 && limit < len(replies.Edges)-startIndex {
		endIndex = startIndex + limit
	}

	var edges []*model.CommentEdge
	for _, reply := range replies.Edges[startIndex:endIndex] {
		if reply.Node != nil {
			edges = append(edges, &model.CommentEdge{
				Cursor: reply.Cursor,
				Node:   reply.Node,
			})
		}
	}

	hasNextPage := endIndex < len(replies.Edges)

	return &model.CommentConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   r.getEndCursorEdges(replies.Edges[startIndex:endIndex]),
			HasNextPage: hasNextPage,
		},
	}, nil
}

func (r *InMemoryRepository) findIndex(cursor string, posts []*model.Post) int {
	for i, post := range posts {
		if post.ID == cursor {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) findCommentIndex(cursor string, comments []*model.Comment) int {
	for i, comment := range comments {
		if comment.ID == cursor {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) findReplyIndex(cursor string, edges []*model.CommentEdge) int {
	for i, edge := range edges {
		if edge.Cursor == cursor {
			return i
		}
	}
	return -1
}

func (r *InMemoryRepository) getEndCursor(posts []*model.Post) *string {
	if len(posts) == 0 {
		return nil
	}
	lastPost := posts[len(posts)-1]
	return &lastPost.ID
}

func (r *InMemoryRepository) getEndCursorComments(comments []*model.Comment) *string {
	if len(comments) == 0 {
		return nil
	}
	lastComment := comments[len(comments)-1]
	return &lastComment.ID
}

func (r *InMemoryRepository) getEndCursorEdges(edges []*model.CommentEdge) *string {
	if len(edges) == 0 {
		return nil
	}
	lastEdge := edges[len(edges)-1]
	return &lastEdge.Cursor
}

func (r *InMemoryRepository) postsToSlice() []*model.Post {
	var posts []*model.Post
	for _, post := range r.posts {
		posts = append(posts, post)
	}
	return posts
}
