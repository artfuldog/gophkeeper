package ui

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// displayUserVerificationPage displays page for user input verification code
// in case of 2-factor authentication.
func (g *Gtui) displayInitSettingsPage(ctx context.Context) {
	selfPage := pageInitSettings

	initUser := g.config.GetUser()
	initSecretKey := g.config.GetSecretKey()
	initServer := g.config.GetServer()
	initMode := g.config.GetMode()
	initSyncInterval := g.config.GetSyncInterval()
	initShowSensitive := g.config.GetShowSensitive()
	initCACert := g.config.GetCACert()

	modes := []string{config.ModeLocal.String(), config.ModeServer.String()}
	curModeIndex := 0

	if initMode == config.ModeServer {
		curModeIndex = 1
	}

	form := tview.NewForm().
		AddInputField("Username", initUser, 40, nil, func(v string) {
			g.config.SetUser(v)
		}).
		AddInputField("Secret Key", initSecretKey, 40, nil, func(v string) {
			g.config.SetSecretKey(v)
		}).
		AddInputField("Server address", initServer, 40, nil, func(v string) {
			g.config.SetServer(v)
		}).
		AddDropDown("Working Mode", modes, curModeIndex, func(option string, optionIndex int) {
			if optionIndex == 0 {
				g.config.SetMode(config.ModeLocal)
				return
			}
			g.config.SetMode(config.ModeServer)
		}).
		AddInputField("Synchronization interval", fmt.Sprint(initSyncInterval), 10, checkFieldInt, func(v string) {
			i, err := (strconv.Atoi(v))
			if err == nil {
				g.config.SetSyncInterval(time.Duration(i) * time.Second)
			}
		}).
		AddCheckbox("Show sensitive", initShowSensitive, func(v bool) {
			g.config.SetShowSensitive(v)
		}).
		AddInputField("CA certificate path", initCACert, 40, nil, func(v string) {
			g.config.SetCACert(v)
		})

	if !g.noConfigFile {
		form.AddButton("Cancel", func() {
			g.config.SetUser(initUser)
			g.config.SetSecretKey(initSecretKey)
			g.config.SetServer(initServer)
			g.config.SetMode(initMode)
			g.config.SetSyncInterval(initSyncInterval)
			g.config.SetShowSensitive(initShowSensitive)
			g.config.SetCACert(initCACert)

			g.pages.RemovePage(selfPage)
			g.displayUserLoginPage(ctx)
			g.setStatus("canceled", 2)
		})
	}

	form.AddButton("Save", func() {
		if err := g.config.Validate(); err != nil {
			g.setStatus(err.Error(), 3)
			return
		}

		if g.noConfigFile {
			if err := g.config.CreateAppDir(); err != nil {
				g.setStatus(err.Error(), 3)
				return
			}
			g.config.SetConfigFile(g.config.GetConfigDefaultFilepath())
		}
		g.noConfigFile = false
		err := g.config.WriteConfig()

		g.pages.RemovePage(selfPage)
		g.displayUserLoginPage(ctx)
		if err != nil {
			g.setStatus(fmt.Sprintf("failed to save settings to '%s': %v", g.config.ConfigFileUsed(), err), 3)
			return
		}
		g.setStatus(fmt.Sprintf("settings successfully saved to '%s'", g.config.ConfigFileUsed()), 2)
	})

	form.SetBorder(true).SetTitle(" Settings ").SetTitleAlign(tview.AlignLeft)
	form.SetButtonsAlign(tview.AlignLeft).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)

	if g.noConfigFile {
		form.SetTitle(" Configure agent settings to access your secret vault ")
	}

	g.pages.AddPage(selfPage, form, true, true)
}

// displayActiveSettingsPage display current settings.
//
// Displayed setting available only for read. For change setting log out required.
func (g *Gtui) displayActiveSettingsPage() {
	selfPage := pageActiveSettings

	form := tview.NewForm().
		AddTextView("Username", g.config.GetUser(), 40, 1, true, false).
		// AddTextView("Secret Key", g.config.GetSecretKey(), 40, 1, true, false).
		AddTextView("Server address", g.config.GetServer(), 40, 1, true, false).
		AddTextView("Working Mode", fmt.Sprint(g.config.GetMode()), 40, 1, true, false)

	if g.config.GetMode() == config.ModeLocal {
		form.AddTextView("Sync interval", fmt.Sprint(g.config.GetSyncInterval()), 10, 1, true, false)
	}

	form.AddTextView("Show sensitive", fmt.Sprint(g.config.GetShowSensitive()), 5, 1, true, false)

	if len(g.config.GetCACert()) > 0 {
		form.AddTextView("CA certificate path", fmt.Sprint(g.config.GetCACert()), 40, 1, true, false)
	}

	form.AddButton("Back", func() {
		g.pages.RemovePage(selfPage)
	})

	form.SetBorder(true).SetTitle(" Settings ").SetTitleAlign(tview.AlignLeft)
	form.SetButtonsAlign(tview.AlignLeft).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)

	form.SetCancelFunc(func() { g.pages.RemovePage(selfPage) })

	g.pages.AddPage(selfPage, form, true, true)
}

// displayMainMenu displays main menu.
// As clientCTX should be passed client's context, as stopClient cancel function for client's context.
// Context and cancel function are used for control status and stops client after user logout.
func (g *Gtui) displayMainMenu(selfCtx, clientCtx context.Context, stopClient context.CancelFunc) {
	selfPage := pageMainMenu

	list := tview.NewList().
		AddItem("Vault", "Browse Vault", 'v', func() {
			g.displayItemBrowser(clientCtx)
		}).
		AddItem("Setting", "Change configuration", 's', g.displayActiveSettingsPage).
		AddItem("About/Help", "About this app", 'a', g.displayAboutHelpMenu).
		AddItem("Log out", "Press to log out", 'l', func() {
			stopClient()

			clientClose := time.NewTimer(api.WaitForClosingInterval)
			select {
			case <-clientClose.C:
			case <-g.clientStopCh:
			}

			g.displayUserLoginPage(selfCtx)
			g.setStatus("logged out", 2)
		}).
		AddItem("Quit", "Press to exit", 'q', g.displayQuitModal)

	list.SetMainTextStyle(tcell.StyleDefault.Bold(true))
	list.SetSecondaryTextColor(tcell.ColorDarkGreen)
	list.SetBorder(true).SetTitle(" GophKeeper ").SetTitleColor(tcell.ColorTomato).
		SetTitleAlign(tview.AlignLeft).SetBorderPadding(1, 0, 2, 2)

	list.SetDoneFunc(g.displayQuitModal)

	g.pages.AddPage(selfPage, list, true, true)
}

// drawMainMenu setups main menu for future use.
func (g *Gtui) displayAboutHelpMenu() {
	selfPage := pageAboutHelp

	about := `
GophKeeper is a client application for accessing vault.
Supported features:
  - Four types of secret items - Login, Card, Note, Data
  - Full CRUD support
  - Two-factor authentication
  - Server-client TLS authentication and encryption 
  - OTP Authenticator Key (for Login)
  - Client side encryption - all data on server stores encrypted`
	help := `
For Navigation use:
  - TAB - next field/button
  - shift+TAB - previous field/button
  - crtl+T / cmd+T - next panel
  - ctrl+Y / cmd+Y - previous panel
  - ESC - go to previous page
`

	form := tview.NewForm().
		AddTextView("About", about, 80, 8, true, false).
		AddTextView("Help", help, 80, 7, true, false).
		AddButton("Back", func() {
			g.pages.RemovePage(selfPage)
		})

	form.SetBorder(true).SetTitle(" About / Help ").SetTitleAlign(tview.AlignLeft)
	form.SetButtonsAlign(tview.AlignLeft).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)

	form.SetCancelFunc(func() { g.pages.RemovePage(selfPage) })

	g.pages.AddPage(selfPage, form, true, true)
}
