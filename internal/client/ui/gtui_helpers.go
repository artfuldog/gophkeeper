package ui

import (
	"context"
	"errors"
	"time"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// checkClientErrors checks returned from client errors:
//   - Session Expiration Error
//
// And defines following actions for known errors.
// Returns true if found known errors and calling function should stops.
func (g *Gtui) checkClientErrorsAndStop(ctx context.Context, err error, parentPage string) bool {
	if errors.Is(err, api.ErrSessionExpired) {
		g.pages.RemovePage(parentPage)
		g.displayUserLoginPage(ctx)
		g.setStatus("session was expired, please log in again...", 5)

		return true
	}

	return false
}

// toPageWithStatus switches to page, sets status text for interval duration.
//
// Status text is optional. If interval is 0 or less sets text status permanently.
func (g *Gtui) toPageWithStatus(pageName string, statusText string, statusInterval int) {
	g.pages.SwitchToPage(pageName)

	if statusText != "" {
		g.setStatus(statusText, statusInterval)
	}
}

// setStatus sets status text for interval duration.
//
// If interval is 0 or less sets text status permanently.
func (g *Gtui) setStatus(text string, interval int) {
	g.status.SetText(text)

	if interval > 0 {
		g.clearStatus(time.Duration(interval) * time.Second)
	}
}

// clearStatus clears text in status string.
func (g *Gtui) clearStatus(interval time.Duration) {
	timer := time.NewTimer(interval)

	go func() {
		<-timer.C
		g.status.Clear()
	}()
}

// captureAndSetFocus is a helper function which configure switching between
// next and previous primitives with provided keys.
//
// Intended for use with SetInputCapture method of tview.Primitive.
func (g *Gtui) captureAndSetFocus(next tview.Primitive, prev tview.Primitive,
	nextKey tcell.Key, prevKey tcell.Key) func(event *tcell.EventKey) *tcell.EventKey { //nolint:unparam

	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == nextKey {
			g.app.SetFocus(next)
			return nil
		}

		if event.Key() == prevKey {
			g.app.SetFocus(prev)
			return nil
		}

		return event
	}
}

// checkFieldInt is input field's checker function which allow insert only digits.
func checkFieldInt(textToCheck string, lastChar rune) bool {
	return !(lastChar < '0' || lastChar > '9')
}
