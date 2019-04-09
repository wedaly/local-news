package controller

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// FeedDetailController handles the UI for details about a particular feed,
// mainly the list of items in the feed.
type FeedDetailController struct {
	appController *AppController
	grid          *tview.Grid
	list          *tview.List
	statusHeader  *tview.TextView
	helpFooter    *tview.TextView
}

func NewFeedDetailController(appController *AppController) PageController {
	// Set up the list of feed items
	list := tview.NewList().
		AddItem("First item", "", 0, nil).
		AddItem("Second item", "", 0, nil).
		ShowSecondaryText(false)
	list.Box.SetBorder(true)

	// Set up a header to display the feed's status (last synced, error, etc)
	statusHeader := tview.NewTextView().
		SetText("Last synced 2020-01-01")

	// Set up a footer to display help text
	const helpText string = `(o) Open in browser   (u) Show/hide unread   (m) Mark all read   (d) Delete   (ESC) Back`
	helpFooter := tview.NewTextView().
		SetText(helpText)

	// Set up a grid to hold the list, header, and footer
	grid := tview.NewGrid().
		SetRows(1, 0, 2).
		AddItem(statusHeader, 0, 0, 1, 1, 0, 0, false).
		AddItem(list, 1, 0, 1, 1, 0, 0, true).
		AddItem(helpFooter, 2, 0, 1, 1, 0, 0, false)

	return &FeedDetailController{appController, grid, list, statusHeader, helpFooter}
}

func (c *FeedDetailController) GetPage() tview.Primitive {
	return c.grid
}

func (c *FeedDetailController) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		c.appController.SwitchToPage(pageFeedList)
		return nil
	}

	return event
}
