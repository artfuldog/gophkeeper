package logger

import (
	"errors"
	"os"

	"github.com/rs/zerolog"
)

const (
	LogTimeFormat = "2006-01-02 15:04:05"
)

// ZLogger represents implemenation of logger bases on famous zerolog package.
type ZLogger struct {
	Logger     zerolog.Logger // logger itself
	LogFile    *os.File       // file for storing logs
	ModuleName string         // module name
}

var _ L = (*ZLogger)(nil)

// NewZLoggerConsole create new ZLogger intstance with console output.
//
// ZLogger allows log message with provided level (severity) or higher, automatically add to log records
// module name.
func NewZLoggerConsole(lvl Level, moduleName string, output OutputFormat) (*ZLogger, error) {
	l := new(ZLogger)
	l.LogFile = nil
	l.ModuleName = moduleName

	zLvl, err := l.GetZeroLogLevel(lvl)
	if err != nil {
		return nil, err
	}

	switch output {
	case OutputStdoutRaw:
		l.Logger = zerolog.New(os.Stdout).Level(zLvl).With().Timestamp().Logger()
	case OutputStdoutPretty:
		l.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: LogTimeFormat}).
			Level(zLvl).With().Timestamp().Logger()
	default:
		return nil, errors.New("unknown output type")
	}

	return l, nil
}

// Trace logs message with Trace level, provided message and component
func (l ZLogger) Trace(message string, component string) {
	l.Logger.Trace().
		Str("module", l.ModuleName).
		Str("component", component).
		Msg(message)
}

// Debug logs message with debug level, provided message and component
func (l ZLogger) Debug(message string, component string) {
	l.Logger.Debug().
		Str("module", l.ModuleName).
		Str("component", component).
		Msg(message)
}

// Info logs message with info level, provided message and component
func (l ZLogger) Info(message string, component string) {
	l.Logger.Info().
		Str("module", l.ModuleName).
		Str("component", component).
		Msg(message)
}

// Warn logs message with warning level, provided error (optional), message and component
func (l ZLogger) Warn(err error, message string, component string) {
	l.Logger.Warn().
		Str("module", l.ModuleName).
		Str("component", component).
		Err(err).
		Msg(message)
}

// Error logs message with error level, provided error (optional), message and component
func (l ZLogger) Error(err error, message string, component string) {
	l.Logger.Error().
		Str("module", l.ModuleName).
		Str("component", component).
		Err(err).
		Msg(message)
}

// Fatal logs message with fatal level, provided error (optional), message and component.
// The os.Exit(1) function is called by the Msg method, which terminates the program immediately.
func (l ZLogger) Fatal(err error, message string, component string) {
	l.Logger.Fatal().
		Str("module", l.ModuleName).
		Str("component", component).
		Err(err).
		Msg(message)
}

// Panic is similiar with Fatal, but instead of os.Exit called panic()
func (l ZLogger) Panic(err error, message string, component string) {
	l.Logger.Panic().
		Str("module", l.ModuleName).
		Str("component", component).
		Err(err).
		Msg(message)
}

// Debug logs message with provided level, error (optional), message and component
func (l ZLogger) Log(level Level, err error, message string, component string) {
	switch level {
	case TraceLevel:
		l.Trace(message, component)
	case DebugLevel:
		l.Debug(message, component)
	case InfoLevel:
		l.Info(message, component)
	case WarnLevel:
		l.Warn(err, message, component)
	case ErrorLevel:
		l.Error(err, message, component)
	case FatalLevel:
		l.Fatal(err, message, component)
	case PanicLevel:
		l.Panic(err, message, component)
	}

}

// GetZeroLogLevel convert logger levels into zerolog levels
func (l ZLogger) GetZeroLogLevel(lvl Level) (zerolog.Level, error) {
	switch lvl {
	case TraceLevel:
		return zerolog.TraceLevel, nil
	case DebugLevel:
		return zerolog.DebugLevel, nil
	case InfoLevel:
		return zerolog.InfoLevel, nil
	case WarnLevel:
		return zerolog.WarnLevel, nil
	case ErrorLevel:
		return zerolog.ErrorLevel, nil
	case FatalLevel:
		return zerolog.FatalLevel, nil
	case PanicLevel:
		return zerolog.PanicLevel, nil
	default:
		return zerolog.NoLevel, errors.New("undefined log level")
	}
}
