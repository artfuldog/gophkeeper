package db

import (
	"testing"

	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := mocklogger.NewMockLogger()

	t.Run("New postgres DB", func(t *testing.T) {
		db, err := New(TypePostgres, &testDBConnParams, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, db)
	})

	t.Run("wrong DB type", func(t *testing.T) {
		_, err := New("wrong type", &testDBConnParams, logger)
		assert.Error(t, err)
	})
}
