package grpcapi

import (
	"errors"
	"fmt"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/server/db"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_wrapErrorToClient(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		message  string
		wantCode codes.Code
	}{
		{
			name:     "Not status code Error",
			err:      assert.AnError,
			wantCode: codes.Unknown,
		},
		{
			name:     "Database ErrNotFound",
			err:      db.ErrNotFound,
			wantCode: codes.NotFound,
		},
		{
			name:     "Database ErrDuplicateEntry",
			err:      db.ErrDuplicateEntry,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Database ErrTransactionFailed",
			err:      db.ErrTransactionFailed,
			wantCode: codes.Internal,
		},
		{
			name:     "Database ErrBadSQLQuery",
			err:      db.ErrBadSQLQuery,
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Database ErrBadSQLQuery with additional info",
			err:      fmt.Errorf("%w::%v", db.ErrBadSQLQuery, db.ErrBadSQLQuery),
			wantCode: codes.InvalidArgument,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := status.Error(tt.wantCode, tt.err.Error())
			if unwErr := errors.Unwrap(tt.err); unwErr != nil {
				want = status.Error(tt.wantCode, unwErr.Error())
			}

			err := wrapErrorToClient(tt.err)
			assert.ErrorIs(t, err, want)

		})
	}
}
