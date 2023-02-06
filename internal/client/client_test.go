package client

import (
	"testing"

	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("Default flags", func(t *testing.T) {
		flags := new(config.Flags)
		client, err := NewClient(flags)
		assert.NoError(t, err)
		assert.NotEmpty(t, client)
	})
	t.Run("Show version", func(t *testing.T) {
		flags := &config.Flags{
			ShowVersion: true,
		}
		agent, err := NewClient(flags)
		assert.NoError(t, err)
		assert.Empty(t, agent)
	})
}

func TestStart(t *testing.T) {
	flags := ReadFlags()

	client, err := NewClient(flags)
	require.NoError(t, err)
	require.NotEmpty(t, client)
}
