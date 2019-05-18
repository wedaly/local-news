package controller

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// PageController handles the UI for a page in the application
// See the tview documentation for details about pages.
type PageController interface {

	// GetPage returns the UI element representing the page
	GetPage() tview.Primitive

	// HandleInput intercepts key events for this page.
	// This is invoked only if the page is currently visible.
	// See the tview documentation for details about event handling.
	HandleInput(event *tcell.EventKey) *tcell.EventKey
}
