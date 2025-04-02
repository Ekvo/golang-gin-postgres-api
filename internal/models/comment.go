// describes - object 'CommentModel' wiht interfaces
package models

import (
	"context"
	"time"
)

type CommentModel struct {
	ID uint

	ArticleID   uint
	ArticleSlug string

	AutorID uint

	CreatedAt time.Time
	UpdatedAt *time.Time

	Body string
}

// CommentNew - save comment in store with return commentID
type CommentNew interface {
	SaveOneComment(ctx context.Context, data any) (uint, error)
}

// CommentFind - read comment from store
type CommentFind interface {
	FindOneComment(ctx context.Context, data any) (CommentModel, error)

	FindCommentList(ctx context.Context, data any) ([]CommentModel, error)
}

// CommentChange -  if exist update comment in store
type CommentChange interface {
	NewDataComment(ctx context.Context, data any) error
}

// CommentRemove - if exist delete comment from store
type CommentRemove interface {
	EndCommentLife(ctx context.Context, data any) error
}
