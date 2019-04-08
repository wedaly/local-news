package controller

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// FeedListController handles the "feed list" page in the UI
type FeedListController struct {
	appController *AppController
	grid          *tview.Grid
	list          *tview.List
	statusHeader  *tview.TextView
	helpFooter    *tview.TextView
}

func NewFeedListController(appController *AppController) PageController {
	// Set up the list of feeds
	list := tview.NewList().
		AddItem("Foo", "", 0, nil).
		AddItem("Bar", "", 0, nil).
		ShowSecondaryText(false)
	list.Box.SetBorder(true).
		SetTitle("All Feeds")

	// Set up the header to display feed loading status
	statusHeader := tview.NewTextView().
		SetText("Loading (1/16)")

	// Set up the footer to show help text
	const helpText string = `(a) Add   (r) Refresh   (u) Show/hide unread   (ESC) Quit`
	helpFooter := tview.NewTextView().
		SetText(helpText)

	// Set up a grid to hold the list, header, and footer
	grid := tview.NewGrid().
		SetRows(1, 0, 2).
		AddItem(statusHeader, 0, 0, 1, 1, 0, 0, false).
		AddItem(list, 1, 0, 1, 1, 0, 0, true).
		AddItem(helpFooter, 2, 0, 1, 1, 0, 0, false)

	// Create the controller and install the handler for list selection events
	c := &FeedListController{appController, grid, list, statusHeader, helpFooter}
	list.SetSelectedFunc(c.handleFeedSelected)

	return c
}

func (c *FeedListController) GetPage() tview.Primitive {
	return c.grid
}

func (c *FeedListController) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Rune() == 'a' {
		c.appController.SwitchToPage(pageAddFeed)
		return nil
	}

	if event.Key() == tcell.KeyEscape {
		c.appController.App.Stop()
		return nil
	}

	return event
}

func (c *FeedListController) handleFeedSelected(int, string, string, rune) {
	// TODO: setup feed detail with feed item info
	c.appController.SwitchToPage(pageFeedDetail)
}
