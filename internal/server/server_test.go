package server

import (
	"context"
	"testing"
	"time"

	"github.com/artfuldog/gophkeeper/internal/mocks/mocklogger"
	"github.com/artfuldog/gophkeeper/internal/server/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	t.Run("Default empty config", func(t *testing.T) {
		cfg, err := NewConfig()
		require.NoError(t, err)
		_, err = NewServer(cfg)
		assert.Error(t, err)
	})

	t.Run("Invalid log level", func(t *testing.T) {
		cfg := new(Config)
		cfg.LogLevel = "unknown level"
		_, err := NewServer(cfg)
		assert.Error(t, err)
	})
}

func TestNewServer_WithDB(t *testing.T) {
	testDBConnParams := db.NewDBParameters("localhost:5432/gophkeeper_db_tests",
		"gksa", "", uint32(50*1024*1024))
	logger := mocklogger.NewMockLogger()
	if _, err := db.New(db.TypePostgres, testDBConnParams, logger); err != nil {
		t.Skipf("SKIPPED. Connetion to local DB failed: %v", err)
		return
	}

	t.Run("Invalid secret key size", func(t *testing.T) {
		cfg := &Config{
			Address:          "127.0.0.1:3200",
			DBType:           "postgres",
			DBDSN:            "localhost:5432/gophkeeper_db_tests",
			DBUser:           "gksa",
			LogLevel:         "fatal",
			MaxSecretSize:    defMaxSecretSize,
			ServerKey:        "123456789f1",
			TokenValidPeriod: defTokenValidPeriod,
		}
		_, err := NewServer(cfg)
		assert.Error(t, err)
	})

	t.Run("TLS certificates missed", func(t *testing.T) {
		cfg := &Config{
			Address:          "127.0.0.1:3200",
			DBType:           "postgres",
			DBDSN:            "localhost:5432/gophkeeper_db_tests",
			DBUser:           "gksa",
			LogLevel:         "fatal",
			MaxSecretSize:    defMaxSecretSize,
			ServerKey:        "123456789f123456789q123456789pQ1",
			TokenValidPeriod: defTokenValidPeriod,
		}
		_, err := NewServer(cfg)
		assert.Error(t, err)
	})

	t.Run("GRPC connection error", func(t *testing.T) {
		cfg := &Config{
			Address:          "127.0.0.1:65636",
			DBType:           "postgres",
			DBDSN:            "localhost:5432/gophkeeper_db_tests",
			DBUser:           "gksa",
			LogLevel:         "fatal",
			MaxSecretSize:    defMaxSecretSize,
			ServerKey:        "123456789f123456789q123456789pQ1",
			TokenValidPeriod: defTokenValidPeriod,
			TLSDisable:       true,
		}
		s, err := NewServer(cfg)
		require.NoError(t, err)
		require.NotEmpty(t, s)

		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan error)
		go s.Run(ctx, ch)

		time.Sleep(2 * time.Second)
		s.DB.Clear(ctx)
		cancel()

		chErr := <-ch
		assert.Error(t, chErr)
	})

	t.Run("Create and start without TLS", func(t *testing.T) {
		cfg := &Config{
			Address:          "127.0.0.1:3200",
			DBType:           "postgres",
			DBDSN:            "localhost:5432/gophkeeper_db_tests",
			DBUser:           "gksa",
			LogLevel:         "fatal",
			MaxSecretSize:    defMaxSecretSize,
			ServerKey:        "123456789f123456789q123456789pQ1",
			TokenValidPeriod: defTokenValidPeriod,
			TLSDisable:       true,
		}
		s, err := NewServer(cfg)
		require.NoError(t, err)
		require.NotEmpty(t, s)

		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan error)
		go s.Run(ctx, ch)

		time.Sleep(2 * time.Second)
		s.DB.Clear(ctx)
		cancel()

		chErr := <-ch
		require.NoError(t, chErr)
	})

	t.Run("Create and start with TLS", func(t *testing.T) {
		cfg := &Config{
			Address:          "127.0.0.1:3200",
			DBType:           "postgres",
			DBDSN:            "localhost:5432/gophkeeper_db_tests",
			DBUser:           "gksa",
			LogLevel:         "fatal",
			MaxSecretSize:    defMaxSecretSize,
			ServerKey:        "123456789f123456789q123456789pQ1",
			TokenValidPeriod: defTokenValidPeriod,
			TLSCertFilepath:  "test_data/service.pem",
			TLSKeyFilepath:   "test_data/service.key",
		}
		s, err := NewServer(cfg)
		require.NoError(t, err)
		require.NotEmpty(t, s)

		ctx, cancel := context.WithCancel(context.Background())
		ch := make(chan error)
		go s.Run(ctx, ch)

		time.Sleep(2 * time.Second)
		s.DB.Clear(ctx)
		cancel()

		chErr := <-ch
		require.NoError(t, chErr)
	})
}
