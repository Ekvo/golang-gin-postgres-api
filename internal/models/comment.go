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

	//maxlenght 2048
	Body string
}

type CommentUpdate interface {
	SaveOneComment(ctx context.Context, data any) (uint, error)

	NewDataComment(ctx context.Context, data any) error

	EndCommentLife(ctx context.Context, data any) error
}

type CommentFind interface {
	FindOneComment(ctx context.Context, data any) (CommentModel, error)
	FindCommentList(ctx context.Context, data any) ([]CommentModel, error)
}

type CommentUpdateFind interface {
	CommentUpdate
	CommentFind
}

type CommentWihtAutor interface {
	CommentUpdateFind
	UserApproveFollowing
}
