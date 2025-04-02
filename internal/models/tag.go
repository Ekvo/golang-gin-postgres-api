// describes - object 'TagModel' wiht interfaces
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

// TagNew - create tag in store with return tagtID
type TagNew interface {
	SaveOneTag(ctx context.Context, data any) (uint, error)
}

// TagRemove - if exist delete tag from store
type TagRemove interface {
	EndTagLife(ctx context.Context, data any) error
}

// TagFind - read comment from store
type TagFind interface {
	FindOneTag(ctx context.Context, data any) (TagModel, error)

	TagsList(ctx context.Context, data any) ([]TagModel, error)
}
