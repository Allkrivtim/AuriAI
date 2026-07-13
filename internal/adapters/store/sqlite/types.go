package sqlite

import (
	"database/sql"

	"github.com/qustavo/dotsql"
)

type Store struct {
	db  *sql.DB
	dot *dotsql.DotSql
}
