package source

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type SQLSource struct {
	sourceDB *sql.DB
}

func NewSQLSource(sourceDB *sql.DB) SQLSource {
	return SQLSource{sourceDB: sourceDB}
}

// SQLRowAndRowsScan - для использования с generic
type SQLRowsRowScan interface {
	*sql.Rows | *sql.Row
	Scan(dest ...any) error
}

// SQLRollback - aborts the transaction
// возвращает ошибку при при откате запроса в базе данных
func SQLRollback(tx *sql.Tx, err error) error {
	if errBack := tx.Rollback(); errBack != nil {
		return fmt.Errorf("query error - %w, tx error - %w", err, errBack)
	}
	return err
}
