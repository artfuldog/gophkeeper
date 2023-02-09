package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
)

// pgErrorsMap is a map listed SQL Erros and corresponds DB Errors.
//
//nolint:gochecknoglobals
var pgErrorsMap = map[string]error{
	"23505": ErrDuplicateEntry,
	"23502": ErrConstraintViolation,
	"23514": ErrConstraintViolation,
	"42703": ErrBadSQLQuery,
	"42601": ErrBadSQLQuery,
}

// wrapPgError is helper function to wrap PostgreSQL's error.
//
// Wraps known type of errors with defined in package errors.
func wrapPgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if dbErr, ok := pgErrorsMap[pgErr.Code]; ok {
			return fmt.Errorf("%w::%v", dbErr, err)
		}
	}

	return stackErrors(ErrUndefinedError, err)
}
