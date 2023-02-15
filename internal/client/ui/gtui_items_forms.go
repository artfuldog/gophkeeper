package ui

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
		URIs:         api.URIs{},
		CustomFields: api.CustomFields{},
	}

	g.pages.AddPage(selfPage, g.drawItemGrid(ctx, item, selfPage, true, true), true, true)
}

// drawItemGrid creates main grid for editing and viewing item.
func (g *Gtui) drawItemGrid(ctx context.Context, item *api.Item, pageName string,
	newItemFlag bool, showSensitive bool) *tview.Grid {

	grid := tview.NewGrid()
	itemMainForm := g.drawItemMainForm(item, pageName, newItemFlag, showSensitive)
	itemInfoForm := g.drawItemInfoForm(item, pageName, newItemFlag)
	itemAdditionsForm := g.drawItemAdditionsForm(ctx, item, pageName, newItemFlag, showSensitive)
	itemButtonsForm := g.drawItemButtonsForm(ctx, item, pageName, newItemFlag, showSensitive)

	grid.SetRows(2, 0, 1).SetColumns(0, 0).
		SetBorders(true).SetBordersColor(tcell.ColorLightSkyBlue).
		AddItem(itemMainForm, 0, 0, 2, 1, 0, 0, true).
		AddItem(itemInfoForm, 0, 1, 1, 1, 0, 0, false).
		AddItem(itemAdditionsForm, 1, 1, 1, 1, 0, 0, false).
		AddItem(itemButtonsForm, 2, 0, 1, 2, 0, 0, false)

	itemMainForm.SetInputCapture(g.captureAndSetFocus(itemButtonsForm, itemAdditionsForm, tcell.KeyCtrlT, tcell.KeyCtrlY))
	itemAdditionsForm.SetInputCapture(g.captureAndSetFocus(itemMainForm, itemButtonsForm, tcell.KeyCtrlT, tcell.KeyCtrlY))
	itemButtonsForm.SetInputCapture(g.captureAndSetFocus(itemAdditionsForm, itemMainForm, tcell.KeyCtrlT, tcell.KeyCtrlY))

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
				AddInputField("CVV", fmt.Sprint(secret.Cvv), 3, checkFieldInt, func(v string) {
					valInt, _ := strconv.Atoi(v)
					secret.Cvv = uint16(valInt)
				})
		} else {
			form.
				AddTextView("Card number", common.MaskLeft(secret.Number, 4), 40, 1, true, false).
				AddTextView("Cardholder", common.MaskLeft(secret.ChName, 5), 40, 1, true, false).
				AddTextView("Expiration Month", common.MaskAll(2), 2, 1, true, false).
				AddTextView("Expiration Year", common.MaskAll(2), 2, 1, true, false).
				AddTextView("CVV", common.MaskAll(3), 3, 1, true, false)
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

// drawItemAdditionsForm creates form for displaying item's custom fields.
func (g Gtui) drawItemAdditionsForm(ctx context.Context, item *api.Item, parentPage string,
	newItemFlag bool, showSensitive bool) *tview.Form {

	form := tview.NewForm()

	selfData := ItemMenuData{
		ParentPage:    parentPage,
		NewItemFlag:   newItemFlag,
		ShowSensitive: showSensitive,
	}

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

				continue
			}

			form.AddTextView(cf.Name, common.MaskAll(8), 40, 1, true, false)
		case common.CfTypeBool:
			form.AddCheckbox(cf.Name, cf.ValueBool, func(v bool) {
				item.CustomFields[index].ValueBool = v
			})
		}
	}

	if item.Type == common.ItemTypeLogin && len(item.URIs) > 0 {
		form.AddTextView("=== URIs:", "", 20, 1, true, false)

		for i, uri := range item.URIs {
			index := i
			form.AddInputField(fmt.Sprintf(" #%d", i), uri.URI, 40, nil, func(v string) {
				item.URIs[index].URI = v
			})
		}
	}

	form.AddButton("Edit custom fields", func() { g.displayCFBrowser(ctx, item, selfData) })

	if item.Type == common.ItemTypeLogin {
		form.AddButton("Edit URIs", func() { g.displayURIBrowser(ctx, item, selfData) })
	}

	form.SetCancelFunc(func() { g.pages.RemovePage(parentPage) })

	form.SetButtonsAlign(tview.AlignLeft).SetBorderPadding(0, 0, 0, 0)
	form.SetButtonStyle(styleCtrButtonsInactive).SetButtonActivatedStyle(styleCtrButtonsActive)

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
