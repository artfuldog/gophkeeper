package ui

import (
	"context"

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

// Primitives styles
//
//nolint:gochecknoglobals
var (
	styleCtrButtonsInactive = tcell.Style{}.Background(tcell.ColorDarkSalmon).Foreground(tcell.ColorBlack)
	styleCtrButtonsActive   = tcell.Style{}.Background(tcell.ColorDarkSlateBlue).Foreground(tcell.ColorWhite)
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
