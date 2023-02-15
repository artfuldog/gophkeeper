package logger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewZLoggerConsole(t *testing.T) {
	_, err := NewZLoggerConsole(100, "module", OutputStdoutRaw)
	require.Error(t, err)

	_, err = NewZLoggerConsole(WarnLevel, "module", 100)
	require.Error(t, err)

	l, err := NewZLoggerConsole(WarnLevel, "module", OutputStdoutRaw)
	require.NoError(t, err)
	assert.NotEmpty(t, l)

	l, err = NewZLoggerConsole(WarnLevel, "module", OutputStdoutPretty)
	require.NoError(t, err)
	assert.NotEmpty(t, l)
}

func TestZLogger_Logging(t *testing.T) {
	logger, _ := NewZLoggerConsole(TraceLevel, "module", OutputStdoutPretty)

	levels := []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel}
	message := "mess"
	component := "component"

	for _, lvl := range levels {
		logger.Log(lvl, assert.AnError, message, component)
	}
}

func TestZLogger_LogPanic(t *testing.T) {
	logger, _ := NewZLoggerConsole(TraceLevel, "module", OutputStdoutPretty)
	message := "mess"
	component := "component"

	assert.Panics(t, func() {
		logger.Log(PanicLevel, assert.AnError, message, component)
	})
}

func TestZLogger_GetZeroLogLevel(t *testing.T) {
	logger, _ := NewZLoggerConsole(TraceLevel, "module", OutputStdoutPretty)

	lvls := map[Level]zerolog.Level{
		TraceLevel: zerolog.TraceLevel,
		DebugLevel: zerolog.DebugLevel,
		InfoLevel:  zerolog.InfoLevel,
		WarnLevel:  zerolog.WarnLevel,
		ErrorLevel: zerolog.ErrorLevel,
		FatalLevel: zerolog.FatalLevel,
		PanicLevel: zerolog.PanicLevel,
	}

	for lvl, zLvl := range lvls {
		got, err := logger.GetZeroLogLevel(lvl)
		require.NoError(t, err)
		assert.Equal(t, got, zLvl)
	}
}
