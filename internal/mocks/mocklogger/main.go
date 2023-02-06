package mocklogger

import (
	"github.com/artfuldog/gophkeeper/internal/logger"
)

type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return new(MockLogger)
}

var _ logger.L = (*MockLogger)(nil)

func (MockLogger) Trace(message string, component string) {}

func (MockLogger) Debug(message string, component string) {}

func (MockLogger) Info(message string, component string) {}

func (MockLogger) Warn(err error, message string, component string) {}

func (MockLogger) Error(err error, message string, component string) {}

func (MockLogger) Fatal(err error, message string, component string) {}

func (MockLogger) Panic(err error, message string, component string) {}

func (MockLogger) Log(level logger.Level, err error, message string, component string) {}
