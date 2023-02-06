package logger

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString_GetLevelFromString(t *testing.T) {
	levels := []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}

	for _, l := range levels {
		got, err := GetLevelFromString(fmt.Sprint(l))
		assert.NoError(t, err)
		assert.Equal(t, l, got)
	}

	got, err := GetLevelFromString(fmt.Sprint(NoLevel))
	require.Error(t, err)
	assert.Equal(t, NoLevel, got)
}
