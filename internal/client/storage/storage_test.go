package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_, err := New("nostor", "user", "path")
	assert.Error(t, err)

	_, err = New(TypeSQLite, "user", "path")
	require.NoError(t, err)
}
