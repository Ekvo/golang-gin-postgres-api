package source

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//var ErrSourceTransaction = errors.New("transaction is rollback")

type poolWithTx struct {
	*pgxpool.Pool
	pgx.Tx
}

func (pt *poolWithTx) Transaction(ctx context.Context, execute func(ctx context.Context) error) (err error) {
	conn, err := pt.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				log.Printf("source: Rollback error - %v", err)
			}
		} else {
			if err := tx.Commit(ctx); err != nil {
				log.Printf("source: Commit error - %v", err)
			}
		}
	}()
	pt.Tx = tx
	err = execute(ctx)
	//if err = execute(ctx); err != nil {
	//	//err = fmt.Errorf("Transaction error - %w", err)
	//	err = ErrSourceTransaction
	//}
	return err
}
