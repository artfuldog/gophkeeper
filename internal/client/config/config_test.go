package config

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()

	os.RemoveAll(appConfigDir)
	os.Exit(exitCode)
}

func TestAgentMode_UnmarshalText(t *testing.T) {
	a := new(AgentMode)

	assert.NoError(t, a.UnmarshalText([]byte("no mode")))
	assert.Equal(t, ModeUnknown, *a)

	assert.NoError(t, a.UnmarshalText([]byte("server")))
	assert.Equal(t, ModeServer, *a)

	assert.NoError(t, a.UnmarshalText([]byte("local")))
	assert.Equal(t, ModeLocal, *a)
}

func TestAgentMode_String(t *testing.T) {
	assert.Equal(t, fmt.Sprint(ModeUnknown), "uknown mode")
	assert.Equal(t, fmt.Sprint(ModeServer), "server")
	assert.Equal(t, fmt.Sprint(ModeLocal), "local")
}

func TestReadFlags(t *testing.T) {
	assert.Equal(t, ReadFlags().ShowVersion, false)
}

func TestNewConfiger(t *testing.T) {
	tests := []struct {
		name    string
		flags   *Flags
		want    *Configer
		wantErr bool
		err     error
	}{
		{
			name:    "New configer without flags",
			flags:   nil,
			want:    &Configer{},
			wantErr: true,
			err:     ErrConfigNotFound,
		},
		{
			name: "Invalid config file path",
			flags: &Flags{
				CustomConfigPath: "./test_data/wrong_config_.yaml",
			},
			want:    &Configer{},
			wantErr: true,
			err:     ErrConfigNotFound,
		},
		{
			name: "Invalid config file",
			flags: &Flags{
				CustomConfigPath: "./test_data/wrong_config.yaml",
			},
			want:    &Configer{},
			wantErr: true,
			err:     assert.AnError,
		},
		{
			name: "Custom config file",
			flags: &Flags{
				CustomConfigPath: "./test_data/custom_config.yaml",
			},
			want:    &Configer{},
			wantErr: false,
			err:     assert.AnError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewConfiger(tt.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfiger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(tt.err, assert.AnError) {
				assert.ErrorIs(t, err, tt.err)
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewConfiger() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestConfiger_Validate(t *testing.T) {
	c := &Configer{Viper: viper.New()}

	assert.ErrorIs(t, ErrEmptyUser, c.Validate())
	c.SetUser("user")
	assert.ErrorIs(t, ErrEmptySecretKey, c.Validate())
	c.SetSecretKey("key")
	assert.ErrorIs(t, ErrEmptyServer, c.Validate())
	c.SetServer("127.0.0.1:1111")
	assert.ErrorIs(t, ErrEmptyAgentMode, c.Validate())
	c.SetMode(ModeLocal)

	assert.NoError(t, c.Validate())
}

func TestConfiger_GettersSetters(t *testing.T) {
	t.Run("Check mode", func(t *testing.T) {
		c := &Configer{Viper: viper.New()}
		c.SetMode(ModeLocal)
		assert.Equal(t, ModeLocal, c.GetMode())

		c.Set("mode", struct{}{})
		assert.Equal(t, ModeUnknown, c.GetMode())

		c.Set("mode", 105)
		assert.Equal(t, ModeUnknown, c.GetMode())
	})

	t.Run("Check Show sensitive", func(t *testing.T) {
		c := &Configer{Viper: viper.New()}
		c.SetShowSensitive(true)
		assert.Equal(t, true, c.GetShowSensitive())
	})

	t.Run("Check log level", func(t *testing.T) {
		c := &Configer{Viper: viper.New()}
		c.SetLogLevel(logger.ErrorLevel)
		assert.Equal(t, logger.ErrorLevel, c.GetLogLevel())

		c.Set("loglevel", 878)
		assert.Equal(t, logger.NoLevel, c.GetLogLevel())

		c.Set("loglevel", "completelywrong")
		assert.Equal(t, logger.NoLevel, c.GetLogLevel())
	})

	t.Run("Check CA Cert", func(t *testing.T) {
		c := &Configer{Viper: viper.New()}
		c.SetCACert("CAcert")
		assert.Equal(t, "CAcert", c.GetCACert())
	})

	t.Run("Check TLS Enable", func(t *testing.T) {
		c := &Configer{Viper: viper.New()}
		c.SetTLSDisable(true)
		assert.Equal(t, true, c.GetTLSDisable())
	})
}

func TestConfiger_CreateAppDir(t *testing.T) {
	c := &Configer{Viper: viper.New()}
	assert.NoError(t, c.CreateAppDir())
}

func TestConfiger_GetConfigDefaultFilepath(t *testing.T) {
	c := &Configer{Viper: viper.New()}
	assert.Equal(t, c.GetConfigDefaultFilepath(), (appConfigDir + appConfigName))
}
