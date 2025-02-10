package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/StonerF/posts/internal/model"
	"github.com/jackc/pgx/v5"
)

type PostgersRep struct {
	db *pgx.Conn
}

func NewPostgresRep(db *pgx.Conn) *PostgersRep {

	return &PostgersRep{db: db}
}

func (db *PostgersRep) CreatePost(authorID, title, content string, allowComments bool) (*model.Post, error) {

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.Insert("posts").Columns("author_id", "title", "content", "allowComments").Values(authorID, title, content, allowComments).Suffix("RETURNING id, author_id, title, content, allow_comments").ToSql()

	if err != nil {
		return nil, fmt.Errorf("Error with query %w", err)
	}
	var post model.Post
	err = db.db.QueryRow(context.Background(), query, args...).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.AllowComments)

	if err != nil {
		return nil, fmt.Errorf("Error with QueryRow %w", err)
	}

	return &post, nil

}

func (db *PostgersRep) GetPosts(limit int, after *string) (*model.PostConnection, error) {

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.Select("id", "author_id", "title", "content", "allow_comments").From("posts").OrderBy("id").Limit(uint64(limit)).ToSql()

	if after != nil {
		query, args, err = psql.Select("id", "author_id", "title", "content", "allow_comments").From("posts").Where("id >", *after).OrderBy("id").Limit(uint64(limit)).ToSql()
	}
	if err != nil {
		return nil, fmt.Errorf("Error with query %w", err)
	}

	rows, err := db.db.Query(context.Background(), query, args...)

	if err != nil {
		return nil, fmt.Errorf("Error with query row %w", err)
	}
	defer rows.Close()

	posts := make([]*model.Post, 0)

	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.AllowComments); err != nil {
			return nil, err
		}
		posts = append(posts, &post)

	}

	edges := make([]*model.PostEdge, len(posts))
	for i, post := range posts {
		edges[i] = &model.PostEdge{
			Cursor: post.ID,
			Node:   post,
		}
	}

	var endCursor *string

	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNextPage := len(posts) == limit

	return &model.PostConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, nil

}

func (db *PostgersRep) GetPostByID(id string) (*model.Post, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query, args, err := psql.Select("id", "author_id", "title", "content", "allow_comments").From("posts").Where("id =", id).ToSql()
	if err != nil {
		return nil, fmt.Errorf("Error with query %w", err)
	}
	var post model.Post
	err = db.db.QueryRow(context.Background(), query, args...).Scan(
		&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.AllowComments,
	)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (db *PostgersRep) CreateComment(authorID, postID string, content string) (*model.Comment, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.Insert("comments").Columns("author_id", "post_id", "content", "created_at").Values(authorID, postID, content, time.Now()).Suffix("RETURNING id, author_id, post_id, content, created_at").ToSql()

	if err != nil {
		return nil, fmt.Errorf("Error with query %w", err)
	}

	var comment model.Comment
	var createdAt time.Time

	err = db.db.QueryRow(context.Background(), query, args...).Scan(
		&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	comment.CreatedAt = createdAt.Format(time.RFC3339)

	return &comment, nil
}

func (db *PostgersRep) GetComments(postID string, limit int, after *string) (*model.CommentConnection, error) {

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	query, args, err := psql.Select("id", "author_id", "post_id", "content", "created_at").From("comments").Where("post_id =", postID).OrderBy("id").Limit(uint64(limit)).ToSql()

	if after != nil {
		query, args, err = psql.Select("id", "author_id", "post_id", "content", "created_at").From("comments").Where("post_id =$1 AND id > $2", postID, *after).OrderBy("id").Limit(uint64(limit)).ToSql()
	}
	if err != nil {
		return nil, fmt.Errorf("Error with query %w", err)
	}

	rows, err := db.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.Comment
	var createdAt time.Time
	for rows.Next() {
		var comment model.Comment
		if err := rows.Scan(&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &createdAt); err != nil {
			return nil, err
		}
		comment.CreatedAt = createdAt.Format(time.RFC3339)
		comments = append(comments, &comment)
	}

	edges := make([]*model.CommentEdge, len(comments))
	for i, comment := range comments {
		edges[i] = &model.CommentEdge{
			Cursor: comment.ID,
			Node:   comment,
		}
	}

	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNextPage := len(comments) == limit

	return &model.CommentConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, nil
}

func (db *PostgersRep) CreateReply(authorID, postID string, content string, parentID *string) (*model.Comment, error) {
	comment, err := db.CreateComment(authorID, postID, content)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO replies_comments (parent_comment_id, reply_comment_id)
		VALUES ($1, $2)
		RETURNING parent_comment_id
		`

	var createdAt time.Time

	err = db.db.QueryRow(context.Background(), query, parentID, comment.ID).Scan(
		&comment.ParentID,
	)
	if err != nil {
		return nil, err
	}

	comment.CreatedAt = createdAt.Format(time.RFC3339)

	return comment, nil
}

func (db *PostgersRep) GetRepliesByCommentID(commentID string, limit int, after *string) (*model.CommentConnection, error) {
	query := `SELECT c.id, c.author_id, c.post_id, rc.parent_comment_id, c.content, c.created_at
			  FROM comments c JOIN replies_comments rc ON c.id = rc.reply_comment_id
			  WHERE rc.parent_comment_id = $1 ORDER BY c.id LIMIT $2`
	args := []interface{}{commentID, limit}
	if after != nil {
		query = `SELECT c.id, c.author_id, c.post_id, rc.parent_comment_id, c.content, c.created_at
				 FROM comments c JOIN replies_comments rc ON c.id = rc.reply_comment_id
				 WHERE rc.parent_comment_id = $1 AND c.id > $3 ORDER BY c.id LIMIT $2`
		args = append(args, *after)
	}

	rows, err := db.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []*model.Comment
	var createdAt time.Time

	for rows.Next() {
		var reply model.Comment
		if err := rows.Scan(&reply.ID, &reply.AuthorID, &reply.PostID, &reply.ParentID, &reply.Content, &createdAt); err != nil {
			return nil, err
		}

		reply.CreatedAt = createdAt.Format(time.RFC3339)
		replies = append(replies, &reply)
	}

	edges := make([]*model.CommentEdge, len(replies))
	for i, reply := range replies {
		edges[i] = &model.CommentEdge{
			Cursor: reply.ID,
			Node:   reply,
		}
	}

	var endCursor *string
	if len(edges) > 0 {
		endCursor = &edges[len(edges)-1].Cursor
	}

	hasNextPage := len(replies) == limit

	return &model.CommentConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
	}, nil
}
