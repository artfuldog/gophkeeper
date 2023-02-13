package ui

import (
	"context"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ItemMenuData represents Item Menu parameters. It used for pass between custom fields
// and URIs form parameters, which required for refreshing item menu after changes to
// custom fields/uri were made.
type ItemMenuData struct {
	ParentPage    string
	NewItemFlag   bool
	ShowSensitive bool
}

// displayCFBrowser displays page listed all item's custom fields.
func (g *Gtui) displayCFBrowser(ctx context.Context, item *api.Item, imData ItemMenuData) {
	selfPage := "Brows"

	browser := tview.NewList()

	backToItemMenuFunc := func() {
		g.pages.RemovePage(selfPage)
		g.pages.AddPage(imData.ParentPage, g.drawItemGrid(ctx, item, imData.ParentPage,
			imData.NewItemFlag, imData.ShowSensitive), true, true)
	}

	r := rune(62)
	for _, cf := range item.CustomFields {
		browser.AddItem(cf.Name, common.CfTypeToText(cf.Type), r, nil)
	}

	browser.SetMainTextStyle(tcell.StyleDefault.Bold(true))
	browser.SetSecondaryTextStyle(tcell.StyleDefault.Italic(true)).
		SetSecondaryTextColor(tcell.ColorDarkGreen)

	title := fmt.Sprintf("  Custom fields of '%s'  ", item.Name)
	browser.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 0, 2, 2).SetTitleColor(tcell.ColorTomato)

	browser.SetSelectedFunc(g.displayEditCfPage(ctx, item, imData))

	browser.SetDoneFunc(backToItemMenuFunc)

	buttons := tview.NewForm().
		AddButton("Add new", func() { g.displayCfCreateModal(ctx, item, imData) }).
		AddButton("Back to item menu", backToItemMenuFunc)

	buttons.SetButtonsAlign(tview.AlignLeft).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive).SetBorderPadding(0, 0, 0, 0)

	browser.SetInputCapture(g.captureAndSetFocus(buttons, buttons, tcell.KeyCtrlT, tcell.KeyCtrlY))
	buttons.SetInputCapture(g.captureAndSetFocus(browser, browser, tcell.KeyCtrlT, tcell.KeyCtrlY))

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(browser, 0, 1, true).AddItem(buttons, 1, 1, false)

	g.pages.AddPage(selfPage, flex, true, true)
}

// displayEditCfPage displays page for edit existing custom field.
func (g *Gtui) displayEditCfPage(ctx context.Context, item *api.Item,
	imData ItemMenuData) func(index int, text, secText string, shortcut rune) {

	return func(index int, text, secText string, shortcut rune) {
		selfPage := pageEditCf

		cfs := item.CustomFields
		cf := cfs[index]

		form := tview.NewForm().SetItemPadding(1).
			AddInputField("Name", cf.Name, 40, nil, func(v string) {
				cf.Name = v
			})

		switch cf.Type {
		case common.CfTypeText:
			form.AddInputField("Value", cf.ValueStr, 40, nil, func(v string) {
				cf.ValueStr = v
			})
		case common.CfTypeHidden:
			form.AddPasswordField("Value", cf.ValueStr, 40, '*', func(v string) {
				cf.ValueStr = v
			})
		case common.CfTypeBool:
			form.AddCheckbox("Value", cf.ValueBool, func(v bool) {
				cf.ValueBool = v
			})
		}

		form.AddButton("Cancel", func() { g.pages.RemovePage(selfPage) }).
			AddButton("Save", func() {
				item.CustomFields[index] = cf
				g.pages.RemovePage(selfPage)
			}).
			AddButton("Delete", func() {
				item.CustomFields = common.DeleteElement(index, cfs)
				g.pages.RemovePage(selfPage)
				g.displayCFBrowser(ctx, item, imData)
			})

		g.pages.AddPage(selfPage, form, true, true)
	}
}

// displayEditCfPage displays page for create custom field.
func (g *Gtui) displayCreateCfPage(ctx context.Context, item *api.Item, cfType string, imData ItemMenuData) {
	selfPage := pageCreateCf

	cf := new(api.CustomField)
	cf.Type = cfType

	form := tview.NewForm().SetItemPadding(1).
		AddInputField("Name", "", 40, nil, func(v string) {
			cf.Name = v
		})

	switch cfType {
	case common.CfTypeText:
		form.AddInputField("Value", "", 40, nil, func(v string) {
			cf.ValueStr = v
		})
	case common.CfTypeHidden:
		form.AddPasswordField("Value", "", 40, '*', func(v string) {
			cf.ValueStr = v
		})
	case common.CfTypeBool:
		form.AddCheckbox("Value", false, func(v bool) {
			cf.ValueBool = v
		})
	}

	form.AddButton("Cancel", func() { g.pages.RemovePage(selfPage) }).
		AddButton("Save", func() {
			item.CustomFields = append(item.CustomFields, *cf)
			g.pages.RemovePage(selfPage)
			g.displayCFBrowser(ctx, item, imData)
		})

	g.pages.AddPage(selfPage, form, true, true)
}

// displayCFBrowser displays page listed all item's URIs.
func (g *Gtui) displayURIBrowser(ctx context.Context, item *api.Item, imData ItemMenuData) {
	selfPage := pageURIBrowser

	browser := tview.NewList()

	backToItemMenuFunc := func() {
		g.pages.RemovePage(selfPage)
		g.pages.AddPage(imData.ParentPage, g.drawItemGrid(ctx, item, imData.ParentPage,
			imData.NewItemFlag, imData.ShowSensitive), true, true)
	}

	r := rune(62)
	for _, uri := range item.URIs {
		browser.AddItem(uri.URI, "", r, nil)
	}

	browser.SetMainTextStyle(tcell.StyleDefault.Bold(true))

	title := fmt.Sprintf("  URIs of '%s'  ", item.Name)
	browser.SetBorder(true).SetTitle(title).SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(1, 0, 2, 2).SetTitleColor(tcell.ColorTomato)

	browser.SetSelectedFunc(g.displayEditURIPage(ctx, item, imData))

	browser.SetDoneFunc(backToItemMenuFunc)

	buttons := tview.NewForm().
		AddButton("Add new", func() { g.displayURIPage(ctx, item, -1, imData) }).
		AddButton("Back to item menu", backToItemMenuFunc)

	buttons.SetButtonsAlign(tview.AlignLeft).SetButtonActivatedStyle(styleCtrButtonsActive).
		SetButtonStyle(styleCtrButtonsInactive).SetBorderPadding(0, 0, 0, 0)

	browser.SetInputCapture(g.captureAndSetFocus(buttons, buttons, tcell.KeyCtrlT, tcell.KeyCtrlY))
	buttons.SetInputCapture(g.captureAndSetFocus(browser, browser, tcell.KeyCtrlT, tcell.KeyCtrlY))

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(browser, 0, 1, true).AddItem(buttons, 1, 1, false)

	g.pages.AddPage(selfPage, flex, true, true)
}

// displayEditCfPage displays page for edit existing URI.
func (g *Gtui) displayEditURIPage(ctx context.Context, item *api.Item,
	imData ItemMenuData) func(index int, text, secText string, shortcut rune) {

	return func(index int, text, secText string, shortcut rune) {
		g.displayURIPage(ctx, item, index, imData)
	}
}

// displayURIPage displays page for create or edit new URI.
//
// Which action to take (create/update) is determiied by index - negative value is a new URI,
// 0 or more - edit existing URI.
func (g *Gtui) displayURIPage(ctx context.Context, item *api.Item, index int, imData ItemMenuData) {
	selfPage := pageURICreateUpdate
	newURI := true

	uri := new(api.URI)
	if index >= 0 {
		*uri = item.URIs[index]
		newURI = false
	}

	form := tview.NewForm().SetItemPadding(1).
		AddInputField("Name", uri.URI, 40, nil, func(v string) {
			uri.URI = v
		}).
		AddInputField("Match", uri.Match, 40, nil, func(v string) {
			uri.Match = v
		})

	form.AddButton("Cancel", func() { g.pages.RemovePage(selfPage) })

	if newURI {
		form.AddButton("Save", func() {
			item.URIs = append(item.URIs, *uri)
			g.pages.RemovePage(selfPage)
			g.displayURIBrowser(ctx, item, imData)
		})
		g.pages.AddPage(selfPage, form, true, true)
		return
	}

	form.AddButton("Save", func() {
		item.URIs[index] = *uri
		g.pages.RemovePage(selfPage)
	}).
		AddButton("Delete", func() {
			item.URIs = common.DeleteElement(index, item.URIs)
			g.pages.RemovePage(selfPage)
			g.displayURIBrowser(ctx, item, imData)
		})

	g.pages.AddPage(selfPage, form, true, true)
}
