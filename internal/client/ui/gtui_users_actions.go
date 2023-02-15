package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/client/api"
)

// userLogin performs user login and starts gRPC client.
func (g *Gtui) userLogin(ctx context.Context, username, password, code string) {
	g.setStatus("Logging in...", 3)

	clientCtx, clientStop := context.WithCancel(ctx)
	g.clientStopCh = make(chan struct{})

	if err := g.client.Connect(clientCtx, g.clientStopCh); err != nil {
		g.setStatus(fmt.Sprintf("failed connect to server: %s", err.Error()), 5)
		clientStop()

		return
	}

	if err := g.client.UserLogin(clientCtx, username, password, code); err != nil {
		if errors.Is(err, api.ErrSecondFactorRequired) {
			g.setStatus("two-factor authentication requested", 0)
			g.displayUserVerificationPage(clientCtx, clientStop, username, password)

			return
		}

		g.setStatus(err.Error(), 5)
		clientStop()

		return
	}

	syncStatusCh := make(chan string)

	if err := g.client.StorageInit(clientCtx, syncStatusCh); err != nil {
		g.setStatus(fmt.Sprintf("forced to server mode - failed to setup local storage: %s", err.Error()), 5)
	}

	go g.bgUpdateSyncStatus(clientCtx, syncStatusCh)

	g.pages.RemovePage(pageUserLogin)
	g.displayMainMenu(ctx, clientCtx, clientStop)
	g.setStatus("Logged in!", 2)
}

// userLogin register user and performs basic checks of user's input.
func (g *Gtui) userRegister(ctx context.Context, user *api.NewUser, parentPage string) {
	g.setStatus("Registering...", 0)

	g.clientStopCh = make(chan struct{})
	if err := g.client.Connect(ctx, g.clientStopCh); err != nil {
		g.setStatus(fmt.Sprintf("failed connect to server: %s", err.Error()), 5)
		return
	}

	if user.Password == "" {
		g.setStatus("password can not be empty", 5)
		return
	}

	if user.Password != user.PasswordConfirm {
		g.setStatus("passwords do not match", 5)
		return
	}

	if user.SecretKey == "" {
		g.setStatus("secret key can not be empty", 5)
		return
	}

	otpOpts, err := g.client.UserRegister(ctx, user)
	if err != nil {
		g.setStatus(err.Error(), 5)
		return
	}

	if user.TwoFactorEnable && otpOpts != nil {
		g.config.SetUser(user.Username)
		g.config.SetSecretKey(user.SecretKey)
		g.displayQRcode(otpOpts, parentPage)

		return
	}

	g.config.SetUser(user.Username)
	g.config.SetSecretKey(user.SecretKey)
	g.pages.RemovePage(parentPage)
	g.toPageWithStatus(pageUserLogin, "Registered", 2)
}
