// Package client implements user-side application for accessing and iteracting with GophKeeer secret vault.
package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/client/ui"
	"github.com/artfuldog/gophkeeper/internal/logger"
)

// Build information
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// Client represents client itself.
type Client struct {
	// API Client for interacting with server
	Client api.Client
	// User interface
	UI ui.UI
}

// ReadFlags is used to read cli's arguments.
func ReadFlags() *config.Flags {
	return config.ReadFlags()
}

// NewClient creates new instance of Client.
func NewClient(flags *config.Flags) (*Client, error) {
	if flags.ShowVersion {
		fmt.Printf("Version: %s\nBuild date: %s\nBuild commit: %s\n",
			buildVersion, buildDate, buildCommit)
		return nil, nil
	}

	var noConfig bool
	configer, err := config.NewConfiger(flags)
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			noConfig = true
		} else {
			return nil, err
		}
	}

	api.GobRegister()

	c := new(Client)

	var clientLogger logger.L
	if clientLogger, err = logger.NewZLoggerConsole(configer.GetLogLevel(),
		"client", logger.OutputStdoutPretty); err != nil {
		return nil, err
	}
	c.Client = api.NewGRPCClient(configer, clientLogger)

	c.UI, _ = ui.New(ui.TypeGtui, c.Client, configer, noConfig)

	return c, err
}

// Start launches UI and starts agent.
func (c *Client) Start(ctx context.Context) error {
	return c.UI.Start(ctx)
}
