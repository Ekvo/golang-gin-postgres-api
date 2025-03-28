package source

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type txBase struct {
	*sql.DB
	*sql.Tx
}

type SQLSource struct {
	sourceDBTX txBase
}

func NewSQLSource(sourceDB *sql.DB) *SQLSource {
	return &SQLSource{sourceDBTX: txBase{DB: sourceDB}}
}

func (tb *txBase) Transaction(ctx context.Context, execute func(ctx context.Context) error) error {
	tx, err := tb.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	tb.Tx = tx
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.Printf("Transaction Rollbak error - %v", err)
		}
	}()
	if err := execute(ctx); err != nil {
		return fmt.Errorf("Transaction error - %w", err)
	}
	return tx.Commit()
}

// SQLRowAndRowsScan - для использования с generic во время сканирования объектов
type SQLRowsRowScan interface {
	*sql.Rows | *sql.Row
	Scan(dest ...any) error
}
