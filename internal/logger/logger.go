// Logger package provides simple way to live logging events and error during app execution.
//
// Package have two implementation of logger:
//   - zlogger wrapping zerolog package
//   - nologger - dummy implementation which implements all methods but nothing does (useful for testing)
//
// Logger provides seven level of events and two output formats in Stdout (console) -
// raw JSON or pretty colored output.
package logger

import (
	"errors"
	"strings"
)

// L represents general Logger interface.
//
// L contains all methods, which particular implementation of logger must implement.
type L interface {
	Log(level Level, err error, message string, component string)
	Trace(message string, component string)
	Debug(message string, component string)
	Info(message string, component string)
	Warn(err error, message string, component string)
	Error(err error, message string, component string)
	Fatal(err error, message string, component string)
	Panic(err error, message string, component string)
}

// Log levels.
type Level int8

// String implements Stringer interface.
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	default:
		return "undefined"
	}
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (l *Level) UnmarshalText(text []byte) (err error) {
	*l, err = GetLevelFromString(string(text))

	return
}

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel

	TraceLevel Level = -1
	NoLevel    Level = -10
)

// Logger output formats.
type OutputFormat int8

const (
	OutputStdoutRaw    OutputFormat = iota // raw JSON
	OutputStdoutPretty                     // pretty formatted plain text
)

func GetLevelFromString(input string) (Level, error) {
	lvlSrt := strings.ToLower(input)

	switch lvlSrt {
	case "trace":
		return TraceLevel, nil
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	default:
		return NoLevel, errors.New("undefined log level")
	}
}
