package controller

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/store"
)

// DeleteSubscriber is notified when a feed is deleted
type DeleteSubscriber interface {
	HandleFeedDeleted(store.FeedId)
}

// DeleteConfirmController is a modal dialog for confirming deletion of a feed
type DeleteConfirmController struct {
	appController *AppController
	feedStore     *store.FeedStore
	modal         *tview.Modal
	feedId        store.FeedId
	subscribers   []DeleteSubscriber
}

func NewDeleteConfirmController(
	appController *AppController,
	feedStore *store.FeedStore) *DeleteConfirmController {

	modal := tview.NewModal().
		AddButtons([]string{"Yes", "No"})

	subscribers := make([]DeleteSubscriber, 0)

	c := &DeleteConfirmController{
		appController,
		feedStore,
		modal,
		store.FeedId(0),
		subscribers,
	}

	modal.SetDoneFunc(c.HandleModalDone)

	return c
}

func (c *DeleteConfirmController) GetPage() tview.Primitive {
	return c.modal
}

func (c *DeleteConfirmController) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	return event
}

// Subscribe registers a subscriber to receive deletion notifications
// This is NOT thread-safe, so it should be called from the main UI thread only.
func (c *DeleteConfirmController) Subscribe(s DeleteSubscriber) {
	c.subscribers = append(c.subscribers, s)
}

// SetFeed sets the feed to be deleted
// This is NOT thread-safe, so it must be called within the UI event loop.
func (c *DeleteConfirmController) SetFeed(feedId store.FeedId) {
	c.feedId = feedId
	c.updateModalText()
}

func (c *DeleteConfirmController) HandleModalDone(buttonIndex int, buttonLabel string) {
	if buttonIndex == 0 {
		c.deleteFeed()
		c.appController.SwitchToPage(pageFeedList)
	} else if buttonIndex == 1 {
		c.appController.SwitchToPage(pageFeedDetail)
	} else {
		panic("Invalid button idx")
	}
}

func (c *DeleteConfirmController) updateModalText() {
	feed, err := c.feedStore.RetrieveFeed(c.feedId)
	if err != nil {
		panic(err)
	}

	confirmText := fmt.Sprintf("Delete feed '%v'?", feed.Name)
	c.modal.SetText(confirmText)
}

func (c *DeleteConfirmController) deleteFeed() {
	if err := c.feedStore.DeleteFeed(c.feedId); err != nil {
		panic(err)
	}

	for _, s := range c.subscribers {
		s.HandleFeedDeleted(c.feedId)
	}
}
