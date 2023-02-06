package authorizer

import (
	"testing"

	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/stretchr/testify/require"
)

func TestYesMan(t *testing.T) {
	logger := mocklogger.NewMockLogger()
	a := NewYesManAuthorizer(logger)
	fields := AuthFields{Username: "username"}

	_, err := a.CreateToken(fields)
	require.NoError(t, err)

	err = a.VerifyToken("", fields)
	require.NoError(t, err)
}
