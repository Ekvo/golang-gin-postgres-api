package models

import (
	"context"
	"time"
)

// TagModel - свойсва тега в базе
type TagModel struct {
	ID uint

	// максимальный размер - 25
	Name string

	AutorID uint

	CreatedAt time.Time
}

type TagUpdate interface {
	SaveOneTag(ctx context.Context, data any) (uint, error)
}

type TagFind interface {
	FindOneTag(ctx context.Context, data any) (TagModel, error)

	TagsList(ctx context.Context, data any) ([]TagModel, error)
}

type TagUpdateFind interface {
	TagUpdate
	TagFind
}

type TagWithAutor interface {
	TagUpdateFind
	UserApproveFollowing
}
