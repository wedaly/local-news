package controller

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/store"
	"github.com/wedaly/local-news/internal/task"
)

// FeedDetailController handles the UI for details about a particular feed,
// mainly the list of items in the feed.
type FeedDetailController struct {
	appController           *AppController
	deleteConfirmController *DeleteConfirmController
	feedStore               *store.FeedStore
	taskManager             *task.TaskManager
	grid                    *tview.Grid
	list                    *tview.List
	statusHeader            *tview.TextView
	helpFooter              *tview.TextView
	feedId                  store.FeedId
}

func NewFeedDetailController(
	appController *AppController,
	deleteConfirmController *DeleteConfirmController,
	feedStore *store.FeedStore,
	taskManager *task.TaskManager) *FeedDetailController {

	// Set up the list of feed items
	list := tview.NewList().
		ShowSecondaryText(false)
	list.Box.SetBorder(true)

	// Set up a header to display the feed's status (last synced, error, etc)
	statusHeader := tview.NewTextView()

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

	c := &FeedDetailController{
		appController,
		deleteConfirmController,
		feedStore,
		taskManager,
		grid,
		list,
		statusHeader,
		helpFooter,
		store.FeedId(0),
	}

	// Subscribe for task updates
	taskManager.Subscribe(c)

	// Subscribe for delete notifications
	deleteConfirmController.Subscribe(c)

	return c
}

func (c *FeedDetailController) GetPage() tview.Primitive {
	return c.grid
}

func (c *FeedDetailController) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		c.appController.SwitchToPage(pageFeedList)
		return nil
	}

	if event.Rune() == 'd' {
		c.deleteConfirmController.SetFeed(c.feedId)
		c.appController.SwitchToPage(pageDeleteConfirm)
		return nil
	}

	return event
}

func (c *FeedDetailController) HandleFeedDeleted(feedId store.FeedId) {
	if c.feedId > 0 && c.feedId == feedId {
		c.feedId = 0
	}
}

// SetDisplayedFeed loads and displayes the latest version of the specified feed
// Assumes that this is called from within the TUI event loop
func (c *FeedDetailController) SetDisplayedFeed(feedId store.FeedId) {
	c.feedId = feedId
	c.LoadFeedDetailsFromStore()
}

func (c *FeedDetailController) LoadFeedDetailsFromStore() {
	feed, err := c.feedStore.RetrieveFeed(c.feedId)
	if err != nil {
		panic(err)
	}

	feedItems, err := c.feedStore.RetrieveFeedItems(c.feedId)
	if err != nil {
		panic(err)
	}

	hasSynced, syncStatus, err := c.feedStore.RetrieveFeedSyncStatus(c.feedId)
	if err != nil {
		panic(err)
	}

	// Display the name of the feed
	boxTitle := fmt.Sprintf("Feed: %v", feed.Name)
	c.list.Box.SetTitle(boxTitle)

	// Replace existing items with items from the database
	c.list.Clear()
	for _, item := range feedItems {
		itemText := fmt.Sprintf("%v  %v", item.Date.Format("2006-01-02"), item.Title)
		c.list.AddItem(itemText, "", 0, nil)
	}

	// Display the feed's last sync status (if any)
	if hasSynced {
		if syncStatus.Success {
			formattedDate := syncStatus.Date.Format("2006-01-02 15:04:05 MST")
			lastSyncedText := fmt.Sprintf("Last synced %v", formattedDate)
			c.statusHeader.SetText(lastSyncedText)
		} else {
			loadErrText := "An error occurred while loading the feed.  Please try reloading the feed later."
			c.statusHeader.SetText(loadErrText)
		}
	} else {
		c.statusHeader.SetText("Loading feed...")
	}
}

func (c *FeedDetailController) HandleTaskScheduled() {
	// ignore
}

func (c *FeedDetailController) HandleTaskCompleted(r task.TaskResult) {
	c.appController.App.QueueUpdateDraw(func() {
		if c.feedId > 0 && c.feedId == r.FeedId {
			c.LoadFeedDetailsFromStore()
		}
	})
}
