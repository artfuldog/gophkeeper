package ui

import (
	"context"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/common"
)

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
