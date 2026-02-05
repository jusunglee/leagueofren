package db

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrNoRows is returned when a query returns no rows
var ErrNoRows = errors.New("no rows in result set")

// IsNoRows returns true if the error indicates no rows were found.
// Works with pgx, database/sql, and the package's own ErrNoRows.
func IsNoRows(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrNoRows) ||
		errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, pgx.ErrNoRows)
}
