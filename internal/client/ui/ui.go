// Package UI implemets User Interface of a client
package ui

import (
	"context"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/client/config"
)

// Supported user interfaces
const (
	TypeGtui = "gtui"
)

// UI represents user interface of agent
type UI interface {
	// Starts user interface.
	Start(context.Context) error
}

// New is a fabric method for create DB with provided type
func New(UIType string, client api.Client, config *config.Configer, noConfig bool) (UI, error) {
	switch UIType {
	case TypeGtui:
		return NewGtui(client, config, noConfig), nil
	default:
		return nil, fmt.Errorf("undefined user interface type: %s", UIType)
	}
}
