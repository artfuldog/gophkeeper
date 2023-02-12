package ui

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"os"
	"strconv"
	"time"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Primitives styles
//
//nolint:gochecknoglobals
var (
	styleCtrButtonsInactive = tcell.Style{}.Background(tcell.ColorDarkSalmon).Foreground(tcell.ColorBlack)
	styleCtrButtonsActive   = tcell.Style{}.Background(tcell.ColorDarkSlateBlue).Foreground(tcell.ColorWhite)
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

// displayItemBrowser displays page listed all user's items sorted alphabetically.
func (g *Gtui) displayItemBrowser(ctx context.Context) {
	selfPage := pageItemBrowser

	items, err := g.client.GetItemsList(ctx)
	if err != nil {
		if g.checkClientErrorsAndStop(ctx, err, selfPage) {
			return
		}

		g.setStatus(err.Error(), 5)

		return
	}

	browser := tview.NewList()

	r := rune(62)
	for _, item := range items {
		browser.AddItem(item.Name, common.ItemTypeText(item.Type), r, nil)
	}

	browser.SetMainTextStyle(tcell.StyleDefault.Bold(true))

	browser.SetSecondaryTextStyle(tcell.StyleDefault.Italic(true)).
		SetSecondaryTextColor(tcell.ColorDarkGreen)

	browser.SetDoneFunc(func() {
		g.pages.RemovePage(selfPage)
	})

	browser.SetSelectedFunc(g.displayEditItemPage(ctx))
	browser.SetBorder(true).SetTitle("  Vault ").SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 0, 2, 2).SetTitleColor(tcell.ColorTomato)

	buttons := tview.NewForm().
		AddButton("Add new Item", func() { g.displayItemCreateModal(ctx) }).
		AddButton("Back to menu", func() { g.pages.RemovePage(selfPage) })

	buttons.SetButtonsAlign(tview.AlignLeft).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive).SetBorderPadding(0, 0, 0, 0)

	browser.SetInputCapture(g.captureAndSetFocus(buttons, buttons, tcell.KeyCtrlT, tcell.KeyCtrlY))
	buttons.SetInputCapture(g.captureAndSetFocus(browser, browser, tcell.KeyCtrlT, tcell.KeyCtrlY))

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(browser, 0, 1, true).AddItem(buttons, 1, 1, false)

	g.pages.AddPage(selfPage, flex, true, true)
}

// displayCreateItemPage displays page for editing existing item.
func (g *Gtui) displayEditItemPage(ctx context.Context) func(index int, text, secText string, shortcut rune) {
	return func(index int, text, secText string, shortcut rune) {
		selfPage := pageItem

		item, err := g.client.GetItem(ctx, text, common.ItemTypeFromText(secText))
		if err != nil {
			if g.checkClientErrorsAndStop(ctx, err, selfPage) {
				return
			}

			g.setStatus(err.Error(), 5)

			return
		}

		g.pages.AddPage(selfPage, g.drawItemGrid(ctx, item, selfPage, false, g.config.GetShowSensitive()), true, true)
	}
}

// displayCreateItemPage displays page for create new item.
func (g *Gtui) displayCreateItemPage(ctx context.Context, itemType string) {
	selfPage := pageItem

	item := &api.Item{
		Type:         itemType,
		Secret:       api.NewSecretEmpty(itemType),
		URIs:         api.URIs{{}},
		CustomFields: api.CustomFields{{}},
	}

	g.pages.AddPage(selfPage, g.drawItemGrid(ctx, item, selfPage, true, true), true, true)
}

// drawItemGrid creates main grid for editing and viewing item.
func (g *Gtui) drawItemGrid(ctx context.Context, item *api.Item, pageName string,
	newItemFlag bool, showSensitive bool) *tview.Grid {

	grid := tview.NewGrid()
	itemMainForm := g.drawItemMainForm(item, pageName, newItemFlag, showSensitive)
	itemInfoForm := g.drawItemInfoForm(item, pageName, newItemFlag)
	itemCFForm := g.drawItemCFForm(item, pageName, showSensitive)
	itemButtonsForm := g.drawItemButtonsForm(ctx, item, pageName, newItemFlag, showSensitive)

	grid.SetRows(2, 0, 1).SetColumns(0, 0).
		SetBorders(true).SetBordersColor(tcell.ColorLightSkyBlue).
		AddItem(itemMainForm, 0, 0, 2, 1, 0, 0, true).
		AddItem(itemInfoForm, 0, 1, 1, 1, 0, 0, false).
		AddItem(itemCFForm, 1, 1, 1, 1, 0, 0, false).
		AddItem(itemButtonsForm, 2, 0, 1, 2, 0, 0, false)

	itemMainForm.SetInputCapture(g.captureAndSetFocus(itemButtonsForm, itemCFForm, tcell.KeyCtrlT, tcell.KeyCtrlY))
	itemCFForm.SetInputCapture(g.captureAndSetFocus(itemMainForm, itemButtonsForm, tcell.KeyCtrlT, tcell.KeyCtrlY))
	itemButtonsForm.SetInputCapture(g.captureAndSetFocus(itemCFForm, itemMainForm, tcell.KeyCtrlT, tcell.KeyCtrlY))

	return grid
}

// drawItemMainForm creates form for displaying item's main editable information.
func (g *Gtui) drawItemMainForm(item *api.Item, pageName string, newItemFlag bool, showSensitive bool) *tview.Form {
	form := tview.NewForm().SetItemPadding(1).
		AddInputField("Name", item.Name, 40, nil, func(v string) {
			item.Name = v
		}).
		AddCheckbox("Reprompt", item.Reprompt, func(v bool) {
			item.Reprompt = v
		})

	switch item.Type {
	case common.ItemTypeLogin:
		secret := item.GetLogin()
		form.AddInputField("Login", secret.Username, 40, nil, func(v string) {
			secret.Username = v
		})

		if showSensitive {
			form.
				AddInputField("Password", secret.Password, 40, nil, func(v string) {
					secret.Password = v
				}).
				AddInputField("Authenticator key", secret.Authkey, 40, nil, func(v string) {
					secret.Authkey = v
				})
		} else {
			curPassword := ""
			if secret.Password != "" {
				curPassword = common.MaskAll(8)
			}
			curAuthKey := ""
			if secret.Authkey != "" {
				curAuthKey = common.MaskAll(8)
			}
			form.
				AddTextView("Password", curPassword, 40, 1, true, false).
				AddTextView("Authenticator key", curAuthKey, 40, 1, true, false)
		}

		if secret.Authkey != "" {
			code, err := crypt.GenerateVerificationCode(secret.Authkey)
			if err != nil {
				code = "invalid authenticator key"
			}

			form.AddTextView("OTP Code", code, 40, 1, true, false)
		}
	case common.ItemTypeCard:
		secret := item.GetCard()

		if showSensitive {
			form.
				AddInputField("Card number", secret.Number, 40, checkFieldInt, func(v string) {
					secret.Number = v
				}).
				AddInputField("Cardholder", secret.ChName, 40, nil, func(v string) {
					secret.ChName = v
				}).
				AddInputField("Expiration Month", fmt.Sprint(secret.ExpMonth), 2, checkFieldInt, func(v string) {
					valInt, _ := strconv.Atoi(v)
					secret.ExpMonth = uint8(valInt)
				}).
				AddInputField("Expiration Year", fmt.Sprint(secret.ExpYear), 2, checkFieldInt, func(v string) {
					valInt, _ := strconv.Atoi(v)
					secret.ExpYear = uint8(valInt)
				}).
				AddInputField("Expiration CVV", fmt.Sprint(secret.Cvv), 3, checkFieldInt, func(v string) {
					valInt, _ := strconv.Atoi(v)
					secret.Cvv = uint16(valInt)
				})
		} else {
			form.
				AddTextView("Card number", common.MaskLeft(secret.Number, 4), 40, 1, true, false).
				AddTextView("Cardholder", common.MaskLeft(secret.ChName, 5), 40, 1, true, false).
				AddTextView("Expiration Month", common.MaskAll(2), 2, 1, true, false).
				AddTextView("Expiration Year", common.MaskAll(2), 2, 1, true, false).
				AddTextView("Expiration CVV", common.MaskAll(3), 3, 1, true, false)
		}
	case common.ItemTypeSecData:
		secret := item.GetSecData()

		form.AddButton("Upload", func() {
			g.displayUploadFileDialog(&secret.Data)
		})

		if !newItemFlag {
			form.AddButton("Download", func() {
				if len(secret.Data) == 0 {
					g.setStatus("no data found", 3)
					return
				}
				g.displayDownloadFileDialog(secret.Data, item.Name)
			})
		}
	}

	form.AddTextArea("Notes", item.Notes, 40, 0, 0, func(v string) {
		item.Notes = v
	})

	form.SetCancelFunc(func() { g.pages.RemovePage(pageName) })
	form.SetBorderPadding(0, 0, 0, 0)

	return form
}

// drawItemInfoForm creates form for displaying item's uneditable information,
// such as type and updated date.
func (g *Gtui) drawItemInfoForm(item *api.Item, pageName string, newItemFlag bool) *tview.Form {
	updated := ""
	if !newItemFlag {
		updated = item.Updated.String()
	}

	form := tview.NewForm().SetItemPadding(0).
		AddTextView("Type", common.ItemTypeText(item.Type), 40, 1, true, false).
		AddTextView("Updated", updated, 40, 1, true, false)

	form.SetCancelFunc(func() { g.pages.RemovePage(pageName) })
	form.SetBorderPadding(0, 0, 0, 0)

	return form
}

// drawItemCFForm creates form for displaying item's custom fields.
func (g Gtui) drawItemCFForm(item *api.Item, pageName string, showSensitive bool) *tview.Form {
	form := tview.NewForm()

	for i, cf := range item.CustomFields {
		index := i

		switch cf.Type {
		case common.CfTypeText:
			form.AddInputField(cf.Name, cf.ValueStr, 40, nil, func(v string) {
				item.CustomFields[index].ValueStr = v
			})
		case common.CfTypeHidden:
			if showSensitive {
				form.AddInputField(cf.Name, cf.ValueStr, 40, nil, func(v string) {
					item.CustomFields[index].ValueStr = v
				})
			} else {
				form.AddPasswordField(cf.Name, cf.ValueStr, 40, '*', func(v string) {
					item.CustomFields[index].ValueStr = v
				})
			}
		case common.CfTypeBool:
			form.AddCheckbox(cf.Name, cf.ValueBool, func(v bool) {
				item.CustomFields[index].ValueBool = v
			})
		}
	}

	form.SetCancelFunc(func() { g.pages.RemovePage(pageName) })
	form.SetBorderPadding(0, 0, 0, 0)

	return form
}

// drawItemButtonsForm creates form with control buttons.
func (g Gtui) drawItemButtonsForm(ctx context.Context, item *api.Item,
	pageName string, newItemFlag bool, showSensitive bool) *tview.Form {

	form := tview.NewForm().
		AddButton("Cancel", func() { g.pages.RemovePage(pageName) }).
		AddButton("Save", func() { g.saveItem(ctx, item, pageName) })

	if showSensitive && !newItemFlag {
		form.AddButton("Hide sensitive", func() {
			g.pages.AddPage(pageName, g.drawItemGrid(ctx, item, pageName, newItemFlag, false), true, true)
		})
	}

	if !showSensitive && !newItemFlag {
		form.AddButton("Show sensitive", func() {
			g.pages.AddPage(pageName, g.drawItemGrid(ctx, item, pageName, newItemFlag, true), true, true)
		})
	}

	if !newItemFlag {
		form.AddButton("Delete", func() { g.deleteItem(ctx, item, pageName) })
	}

	form.SetCancelFunc(func() { g.pages.RemovePage(pageName) })
	form.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)
	form.SetButtonStyle(styleCtrButtonsInactive).SetButtonActivatedStyle(styleCtrButtonsActive)

	return form
}

// displayOTPKey displays TOTP Secret key for setup 2-factor authentication.
func (g *Gtui) displayUploadFileDialog(itemData *[]byte) {
	selfPage := pageUploadFile

	var filepath string

	form := tview.NewForm().
		AddInputField("File path", "", 70, nil, func(v string) {
			filepath = v
		}).
		AddButton("Upload", func() {
			var err error
			*itemData, err = os.ReadFile(filepath)
			if err != nil {
				g.pages.RemovePage(selfPage)
				g.setStatus(err.Error(), 3)
				return
			}
			g.pages.RemovePage(selfPage)
			g.setStatus(fmt.Sprintf("File '%s' is successfully uploaded", filepath), 3)
		}).
		AddButton("Cancel", func() { g.pages.RemovePage(selfPage) })

	form.SetBorder(true).SetBackgroundColor(tcell.ColorDarkBlue).
		SetTitle(" Upload file ").SetTitleAlign(tview.AlignCenter)
	form.SetButtonsAlign(tview.AlignCenter).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)
	form.SetCancelFunc(func() { g.pages.RemovePage(selfPage) })

	grid := tview.NewGrid().
		SetColumns(0, 60, 0).SetRows(0, 7, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)

	g.pages.AddPage(selfPage, grid, true, true)
}

// displayOTPKey displays TOTP Secret key for setup 2-factor authentication.
func (g *Gtui) displayDownloadFileDialog(itemData []byte, itemName string) {
	selfPage := pageDownloadFile

	path, _ := os.Executable()
	filepath := fmt.Sprintf("%s/%s", path, itemName)

	form := tview.NewForm().
		AddInputField("File path", filepath, 70, nil, func(v string) {
			filepath = v
		}).
		AddButton("Download", func() {
			err := os.WriteFile(filepath, itemData, 06)
			if err != nil {
				g.pages.RemovePage(selfPage)
				g.setStatus(err.Error(), 3)
				return
			}
			g.pages.RemovePage(selfPage)
			g.setStatus(fmt.Sprintf("File is successfully downloaded as '%s'", filepath), 3)
		}).
		AddButton("Cancel", func() { g.pages.RemovePage(selfPage) })

	form.SetBorder(true).SetBackgroundColor(tcell.ColorDarkBlue).
		SetTitle(" Download file ").SetTitleAlign(tview.AlignCenter)
	form.SetButtonsAlign(tview.AlignCenter).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive)
	form.SetCancelFunc(func() { g.pages.RemovePage(selfPage) })

	grid := tview.NewGrid().
		SetColumns(0, 60, 0).SetRows(0, 7, 0).
		AddItem(form, 1, 1, 1, 1, 0, 0, true)

	g.pages.AddPage(selfPage, grid, true, true)
}
