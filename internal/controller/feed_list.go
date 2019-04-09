package controller

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/store"
)

// FeedListController handles the "feed list" page in the UI
type FeedListController struct {
	appController        *AppController
	feedDetailController *FeedDetailController
	feedStore            *store.FeedStore
	feedRecords          []store.FeedRecord
	grid                 *tview.Grid
	list                 *tview.List
	statusHeader         *tview.TextView
	helpFooter           *tview.TextView
}

func NewFeedListController(
	appController *AppController,
	feedDetailController *FeedDetailController,
	feedStore *store.FeedStore) *FeedListController {

	// Set up the list of feeds
	list := tview.NewList().
		ShowSecondaryText(false)
	list.Box.SetBorder(true).
		SetTitle("All Feeds")

	// Set up the header to display feed loading status
	statusHeader := tview.NewTextView().
		SetText("Refreshing 10 feeds")

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

	// Create slice for holding feed records
	feedRecords := make([]store.FeedRecord, 0)

	// Create the controller and install the handler for list selection events
	c := &FeedListController{
		appController,
		feedDetailController,
		feedStore,
		feedRecords,
		grid,
		list,
		statusHeader,
		helpFooter,
	}
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

func (c *FeedListController) LoadFeedsFromStore() {
	feedRecords, err := c.feedStore.RetrieveFeeds()
	if err != nil {
		panic(err)
	}

	// Replace existing items with feeds from the database
	c.list.Clear()
	for _, feed := range feedRecords {
		c.list.AddItem(feed.Name, "", 0, nil)
	}

	// Store the feed records in memory
	// so we can retrieve them later using the list item index
	c.feedRecords = feedRecords
}

func (c *FeedListController) handleFeedSelected(idx int, text string, secondaryText string, shortcut rune) {
	feed := c.feedRecords[idx]
	c.feedDetailController.SetDisplayedFeed(feed.Id, feed.Name)
	c.appController.SwitchToPage(pageFeedDetail)
}
