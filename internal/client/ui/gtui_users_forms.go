package ui

import (
	"bytes"
	"context"
	"image/png"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// displayLoginMenu displays initial login page.
func (g *Gtui) displayUserLoginPage(ctx context.Context) {
	selfPage := pageUserLogin

	var password string
	form := tview.NewForm().
		AddInputField("Username", g.config.GetUser(), 41, nil, func(v string) {
			g.config.SetUser(v)
		}).
		AddPasswordField("Password", password, 41, '*', func(v string) {
			password = v
		}).
		AddButton("Login", func() {
			g.userLogin(ctx, g.config.GetUser(), password, "")
		}).
		AddButton("Settings", func() {
			g.displayInitSettingsPage(ctx)
		}).
		AddButton("Create account", func() {
			g.displayRegisterPage(ctx)
		}).
		AddButton("Quit", g.displayQuitModal)

	form.SetBorder(true).SetTitle(" Log in or create a new account to access your secure vault ").
		SetTitleAlign(tview.AlignLeft)

	g.pages.AddPage(selfPage, form, true, true)
}

// displayUserVerificationPage displays page for user input verification code
// in case of 2-factor authentication. As stopClient should be passed cancel function for
// client's context.
func (g *Gtui) displayUserVerificationPage(ctx context.Context, stopClient context.CancelFunc,
	username, password string) {

	selfPage := pageUserVerification

	var code string
	form := tview.NewForm().
		AddInputField("Code", code, 15, nil, func(v string) {
			code = v
		}).
		AddButton("Done", func() {
			g.pages.RemovePage(selfPage)
			g.userLogin(ctx, username, password, code)
		}).
		AddButton("Cancel", func() {
			stopClient()
			g.clearStatus(0)
			g.pages.RemovePage(selfPage)
		})

	form.SetBorder(true).SetBackgroundColor(tcell.ColorDarkBlue).
		SetTitle(" Enter verification code ").SetTitleAlign(tview.AlignCenter)
	form.SetButtonsAlign(tview.AlignCenter).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)

	grid := tview.NewGrid().
		SetColumns(0, 30, 0).SetRows(0, 7, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)

	g.pages.AddPage(selfPage, grid, true, true)
}

// displayRegisterPage displays user registration page.
func (g *Gtui) displayRegisterPage(ctx context.Context) {
	selfPage := pageUserRegister
	newUser := new(api.NewUser)

	newUser.Username = g.config.GetUser()

	form := tview.NewForm().
		AddInputField("Username", newUser.Username, 25, nil, func(v string) {
			newUser.Username = v
		}).
		AddPasswordField("Password", newUser.Password, 25, '*', func(v string) {
			newUser.Password = v
		}).
		AddPasswordField("Confirm Password", newUser.PasswordConfirm, 25, '*', func(v string) {
			newUser.PasswordConfirm = v
		}).
		AddInputField("Secret key", newUser.SecretKey, 25, nil, func(v string) {
			newUser.SecretKey = v
		}).
		AddInputField("E-mail", newUser.Email, 25, nil, func(v string) {
			newUser.Email = v
		}).
		AddCheckbox("Two-Factor Authentication", newUser.TwoFactorEnable, func(v bool) {
			newUser.TwoFactorEnable = v
		}).
		AddButton("Cancel", func() { g.pages.RemovePage(selfPage) }).
		AddButton("Register", func() { g.userRegister(ctx, newUser, selfPage) })

	form.SetCancelFunc(func() { g.pages.RemovePage(selfPage) })

	form.SetBorder(true).SetTitle(" New user registration ").
		SetTitleAlign(tview.AlignLeft)

	g.pages.AddPage(selfPage, form, true, true)
}

// displayQRcode displays QRCode for setup 2-factor authentication.
func (g *Gtui) displayQRcode(otp *api.TOTPKey, parentPage string) {
	selfPage := pageQRCode

	image := tview.NewImage()
	qrcode, _ := png.Decode(bytes.NewReader(otp.QRCode))
	image.SetImage(qrcode)

	form := tview.NewForm().
		AddButton("OK", func() {
			g.clearStatus(0)
			g.pages.RemovePage(selfPage)
			g.displayOTPKey(otp.SecretKey, parentPage)
		})
	form.SetButtonsAlign(tview.AlignCenter).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)

	grid := tview.NewGrid().
		SetColumns(0, 300, 0).SetRows(0, 300, 3).
		AddItem(image, 1, 1, 1, 1, 0, 0, false).
		AddItem(form, 2, 1, 1, 1, 0, 0, true)

	g.pages.AddPage(selfPage, grid, true, true)
}

// displayOTPKey displays TOTP Secret key for setup 2-factor authentication.
func (g *Gtui) displayOTPKey(key string, parentPage string) {
	selfPage := pageOTPCode

	form := tview.NewForm().
		AddTextView("", key, 50, 1, true, false).
		AddButton("OK", func() {
			g.clearStatus(0)
			g.pages.RemovePage(selfPage)
			g.pages.RemovePage(parentPage)
			g.toPageWithStatus(pageUserLogin, "Registered", 2)
		})

	form.SetBorder(true).SetBackgroundColor(tcell.ColorDarkBlue).
		SetTitle(" Your authenticator key ").SetTitleAlign(tview.AlignCenter)
	form.SetButtonsAlign(tview.AlignCenter).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)
	form.SetBorderPadding(0, 0, 0, 0)

	grid := tview.NewGrid().
		SetColumns(0, 40, 0).SetRows(0, 7, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)

	g.pages.AddPage(selfPage, grid, true, true)
}
