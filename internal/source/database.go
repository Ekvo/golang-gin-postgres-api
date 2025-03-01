package source

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type SQLSource struct {
	sourceDB *sql.DB
}

func NewSQLSource(sourceDB *sql.DB) SQLSource {
	return SQLSource{sourceDB: sourceDB}
}
