package controller

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/store"
)

// FeedDetailController handles the UI for details about a particular feed,
// mainly the list of items in the feed.
type FeedDetailController struct {
	appController *AppController
	feedStore     *store.FeedStore
	grid          *tview.Grid
	list          *tview.List
	statusHeader  *tview.TextView
	helpFooter    *tview.TextView
}

func NewFeedDetailController(appController *AppController, feedStore *store.FeedStore) *FeedDetailController {
	// Set up the list of feed items
	list := tview.NewList().
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

	return &FeedDetailController{
		appController,
		feedStore,
		grid,
		list,
		statusHeader,
		helpFooter,
	}
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

// SetDisplayedFeed loads and displayes the latest version of the specified feed
func (c *FeedDetailController) SetDisplayedFeed(feedId store.FeedId) {
	feed, err := c.feedStore.RetrieveFeed(feedId)
	if err != nil {
		panic(err)
	}

	feedItems, err := c.feedStore.RetrieveFeedItems(feedId)
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
}
