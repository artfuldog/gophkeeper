package authorizer

import (
	"testing"
	"time"

	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger := mocklogger.NewMockLogger()
	tokenDur := 5 * time.Minute

	a, err := New(TypeJWT, "123456789a123456789a123456789abc", tokenDur, logger)
	require.Error(t, err)
	assert.Empty(t, a)

	a, err = New(TypePaseto, "123456789a123456789a123456789abc", tokenDur, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, a)

	a, err = New(TypeYesMan, "123456789a123456789a123456789abc", tokenDur, logger)
	require.NoError(t, err)
	assert.NotEmpty(t, a)

	a, err = New("notype", "123456789a123456789a123456789abc", tokenDur, logger)
	require.Error(t, err)
	assert.Empty(t, a)
}
