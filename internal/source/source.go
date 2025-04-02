package source

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrSourceAlreadyExists - во время регистрации(signup)
var ErrSourceAlreadyExists = errors.New("resource already exists")

var ErrSourceNotFound = errors.New("resource not found")

type SQLSource struct {
	pTx poolWithTx
}

func NewSQLSource(pool *pgxpool.Pool) SQLSource {
	return SQLSource{pTx: poolWithTx{Pool: pool}}
}

func (s SQLSource) NewTable(ctx context.Context, data ...any) error {
	createTable := func(ctx context.Context) error {
		for _, table := range data {
			_, err := s.pTx.Tx.Exec(ctx, table.(string))
			if err != nil {
				return err
			}
		}
		return nil
	}
	return s.pTx.Transaction(ctx, createTable)
}
