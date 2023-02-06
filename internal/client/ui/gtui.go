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

// Page names
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
	app    *tview.Application
	status *tview.TextView
	pages  *tview.Pages

	client       api.Client
	config       *config.Configer
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

	g.pages = tview.NewPages()

	clientCtx, clientStop := context.WithCancel(ctx)
	defer clientStop()

	g.drawMainMenu(clientCtx)

	if g.noConfigFile {
		g.displayInitSettingsPage(clientCtx)
	} else {
		g.displayUserLoginPage(clientCtx)
	}

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(g.pages, 0, 1, true).
			AddItem(g.status, 3, 1, false), 0, 1, true)

	if err := g.app.SetRoot(flex, true).EnableMouse(true).SetFocus(flex).Run(); err != nil {
		return err
	}

	return nil
}

// checkClientErrors checks returned from client errors:
//  - Session Expiration Error
// And defines following actions for known errors.
// Returns true if found known erros and calling function should stops.
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
		g.clearStatus(time.Duration(interval))
	}
}

// clearStatus clears text in status string.
func (g *Gtui) clearStatus(interval time.Duration) {
	timer := time.NewTimer(interval * time.Second)
	go func() {
		<-timer.C
		g.status.Clear()
	}()
}

// captureAndSetFocus is a capcture function which configure switching between
// next and prevous primitives with provided keys.
//
// Intended for use with SetInputCapture method of tview.Primitive.
func (g *Gtui) captureAndSetFocus(next tview.Primitive, prev tview.Primitive,
	nextKey tcell.Key, prevKey tcell.Key) func(event *tcell.EventKey) *tcell.EventKey {

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
