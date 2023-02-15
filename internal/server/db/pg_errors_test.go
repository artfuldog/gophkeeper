package db

import (
	"testing"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestWrapPgError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name: "Nil error",
		},
		{
			name:    "Undefined error",
			err:     assert.AnError,
			wantErr: ErrUndefinedError,
		},
		{
			name: "Known PG error",
			err: &pgconn.PgError{
				Code: "23505",
			},
			wantErr: ErrDuplicateEntry,
		},
		{
			name: "Unknown PG error",
			err: &pgconn.PgError{
				Code: "99999",
			},
			wantErr: ErrUndefinedError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, wrapPgError(tt.err), tt.wantErr)
		})
	}
}
