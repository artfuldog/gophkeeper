package ui

import (
	"context"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/gdamore/tcell/v2"
)

// bgUpdateSyncStatus is a background function for update sync status field.
func (g *Gtui) bgUpdateSyncStatus(ctx context.Context, ch <-chan string) {
	for {
		select {
		case <-ctx.Done():
			g.clearSyncStatus()
			return
		case status := <-ch:
			if status == api.SyncStatusFailed {
				g.syncStatus.SetBorder(true).SetBorderColor(tcell.ColorDarkRed)
				g.syncStatus.SetText(status)

				continue
			}

			g.syncStatus.SetBorder(true).SetBorderColor(tcell.ColorDarkGreen)
			g.syncStatus.SetText(status)
		}
	}
}

// clearSyncStatus is a helper function for clearing sync status field.
func (g *Gtui) clearSyncStatus() {
	g.syncStatus.SetBorder(true).SetBorderColor(tcell.ColorDarkGreen)
	g.syncStatus.Clear()
}
