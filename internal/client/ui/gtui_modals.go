package ui

import (
	"context"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/rivo/tview"
)

// displayQuitModal displays modal exit app window.
func (g *Gtui) displayQuitModal() {
	selfPage := modalQuit

	modal := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{"Cancel", "Quit"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				g.app.Stop()
			}
			if buttonLabel == "Cancel" {
				g.pages.RemovePage(selfPage)
				g.setStatus("Canceled", 2)
			}
		})

	g.setStatus("Wait for user confirmation...", 0)
	g.pages.AddPage(selfPage, modal, true, true)
}

// displayItemCreateModal displays modal window with available item's types,
// reads user input, creates and switches to item create page.
func (g *Gtui) displayItemCreateModal(ctx context.Context) {
	selfPage := modalItemType

	modal := tview.NewModal().
		SetText("Choose item type").
		AddButtons([]string{"Login", "Card", "Note", "Data"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			var itemType string

			switch buttonLabel {
			case "Login":
				itemType = common.ItemTypeLogin
			case "Card":
				itemType = common.ItemTypeCard
			case "Note":
				itemType = common.ItemTypeSecNote
			case "Data":
				itemType = common.ItemTypeSecData
			default:
				g.pages.RemovePage(selfPage)
				g.setStatus("canceled...", 2)
				return
			}

			g.pages.RemovePage(selfPage)
			g.setStatus(fmt.Sprintf("creating new %s item", common.ItemTypeText(itemType)), 2)
			g.displayCreateItemPage(ctx, itemType)
		})

	g.setStatus("Wait for user input...", 0)
	g.pages.AddPage(selfPage, modal, true, true)
}
