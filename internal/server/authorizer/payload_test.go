package authorizer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPaylodValid(t *testing.T) {
	p, err := NewPayload("username", 5*time.Second)
	require.NoError(t, err)

	p.ExpiredAt = time.Now().Add(-10 * time.Minute)

	fields := AuthFields{Username: "username"}
	err = p.Valid(fields)
	require.ErrorIs(t, err, ErrExpiredToken)
}
