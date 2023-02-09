package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/common"
)

// userLogin performs user login and starts gRPC client.
func (g *Gtui) userLogin(ctx context.Context, username, password, code string) {
	g.setStatus("Logging in...", 3)

	if err := g.client.Connect(ctx); err != nil {
		g.setStatus(fmt.Sprintf("failed connect to server: %s", err.Error()), 5)

		return
	}

	if err := g.client.UserLogin(ctx, username, password, code); err != nil {
		if errors.Is(err, api.ErrSecondFactorRequired) {
			g.setStatus("two-factor authentication requested", 0)
			g.displayUserVerificationPage(ctx, username, password)

			return
		}

		g.setStatus(err.Error(), 5)

		return
	}

	g.pages.RemovePage(pageUserLogin)
	g.toPageWithStatus(pageMainMenu, "Logged in!", 2)
}

// userLogin register user and performs basic checks of user's input.
func (g *Gtui) userRegister(ctx context.Context, user *api.NewUser, parentPage string) {
	g.setStatus("Registering...", 0)

	if err := g.client.Connect(ctx); err != nil {
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

// saveItem saves item.
func (g *Gtui) saveItem(ctx context.Context, item *api.Item, pageName string) {
	if err := g.client.SaveItem(ctx, item); err != nil {
		if g.checkClientErrorsAndStop(ctx, err, pageName) {
			return
		}

		g.setStatus(err.Error(), 3)

		return
	}

	g.pages.RemovePage(pageName)
	g.displayItemBrowser(ctx)
}

// deleteItem delete item.
func (g *Gtui) deleteItem(ctx context.Context, item *api.Item, pageName string) {
	if err := g.client.DeleteItem(ctx, item); err != nil {
		if g.checkClientErrorsAndStop(ctx, err, pageName) {
			return
		}

		g.setStatus(err.Error(), 5)

		return
	}

	g.pages.RemovePage(pageName)
	g.displayItemBrowser(ctx)
	g.setStatus(fmt.Sprintf("item '%s' (%s) was deleted", item.Name, common.ItemTypeText(item.Type)), 5)
}
