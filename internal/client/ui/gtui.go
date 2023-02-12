package ui

import (
	"context"
	"errors"
	"time"

	"github.com/artfuldog/gophkeeper/internal/client/api"
	"github.com/artfuldog/gophkeeper/internal/client/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Page names.
const (
	pageMainMenu         = "Main menu"
	pageUserLogin        = "User login"
	pageUserVerification = "User verification"
	pageUserRegister     = "User register"
	pageQRCode           = "QRCode page"
	pageOTPCode          = "OTP code page"
	pageInitSettings     = "Settings page"
	pageActiveSettings   = "Active Settings page"
	pageAboutHelp        = "About Help page"
	pageItemBrowser      = "Item browser"
	pageItem             = "Item page"
	pageUploadFile       = "Upload File"
	pageDownloadFile     = "Dpwnload File"

	modalQuit     = "Quit modal"
	modalItemType = "Item Type Modal"
)

// Gtui represents graphical terminal user interface.
type Gtui struct {
	// Main instance.
	app *tview.Application
	// Text view in the bottom for displaying notifications for user.
	status *tview.TextView
	// Text view in the right bottom corner for displaying synchronization status.
	syncStatus *tview.TextView
	// Main widget.
	pages *tview.Pages
	// Client provides API for interactions with server and local storage (optional).
	client api.Client
	// Channel for receiving notifications from client about its stop.
	clientStopCh chan struct{}
	// Configer instance for read and change configuraion parameters live.
	config *config.Configer
	// Flag indicates that no configuration was found. Used when gtui starts for displayng
	// setting page.
	noConfigFile bool
}

var _ UI = (*Gtui)(nil)

// NewGTUIClient creates a tui client with all configs and dependencies needed.
func NewGtui(client api.Client, config *config.Configer, noConfig bool) *Gtui {
	return &Gtui{
		client:       client,
		config:       config,
		noConfigFile: noConfig,
	}
}

// Start starts graphical terminal user interface.
func (g *Gtui) Start(ctx context.Context) error {
	g.app = tview.NewApplication()

	g.status = tview.NewTextView()
	g.status.SetBorder(true).SetBorderColor(tcell.ColorDarkGreen)
	g.status.SetTextStyle(tcell.StyleDefault.Blink(true))

	g.syncStatus = tview.NewTextView()
	g.syncStatus.SetBorder(true).SetBorderColor(tcell.ColorDarkGreen)
	g.syncStatus.SetTextStyle(tcell.StyleDefault.Blink(true))
	g.syncStatus.SetTextAlign(tview.AlignCenter)

	g.pages = tview.NewPages()

	if g.noConfigFile {
		g.displayInitSettingsPage(ctx)
	} else {
		g.displayUserLoginPage(ctx)
	}

	statusFlex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(g.status, 0, 1, true).
			AddItem(g.syncStatus, 14, 1, false), 0, 1, true)

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(g.pages, 0, 1, true).
			AddItem(statusFlex, 3, 1, false), 0, 1, true)

	appStopCh := make(chan error)

	go func() {
		appStopCh <- g.app.SetRoot(flex, true).EnableMouse(true).SetFocus(flex).Run()
	}()

	select {
	case <-ctx.Done():
		g.app.Stop()
		return <-appStopCh
	case err := <-appStopCh:
		return err
	}
}

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
