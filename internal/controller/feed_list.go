package controller

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/store"
	"github.com/wedaly/local-news/internal/task"
)

// FeedListController handles the "feed list" page in the UI
type FeedListController struct {
	appController        *AppController
	feedDetailController *FeedDetailController
	feedStore            *store.FeedStore
	taskManager          *task.TaskManager
	grid                 *tview.Grid
	list                 *tview.List
	statusHeader         *tview.TextView
	helpFooter           *tview.TextView
	listIdxToFeedId      []store.FeedId
	numUncompletedTasks  int
}

func NewFeedListController(
	appController *AppController,
	feedDetailController *FeedDetailController,
	feedStore *store.FeedStore,
	taskManager *task.TaskManager) *FeedListController {

	// Set up the list of feeds
	list := tview.NewList().
		ShowSecondaryText(false)
	list.Box.SetBorder(true).
		SetTitle("All Feeds")

	// Set up the header to display feed loading status
	statusHeader := tview.NewTextView()

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
	listIdxToFeedId := make([]store.FeedId, 0)

	// Create the controller and install the handler for list selection events
	c := &FeedListController{
		appController,
		feedDetailController,
		feedStore,
		taskManager,
		grid,
		list,
		statusHeader,
		helpFooter,
		listIdxToFeedId,
		0,
	}
	list.SetSelectedFunc(c.handleFeedSelected)

	// Subscribe for task updates
	taskManager.Subscribe(c)

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

	if event.Rune() == 'r' {
		c.RefreshAllFeeds()
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

	// Replace existing items with feeds from the database.
	// Keep track of the feed ID for each item in the list
	// so we can operate on them later.
	c.list.Clear()
	c.listIdxToFeedId = make([]store.FeedId, len(feedRecords))
	for i, feed := range feedRecords {
		c.list.AddItem(feed.Name, "", 0, nil)
		c.listIdxToFeedId[i] = feed.Id
	}
}

func (c *FeedListController) RefreshAllFeeds() {
	if c.numUncompletedTasks > 0 {
		return
	}

	for _, feedId := range c.listIdxToFeedId {
		c.taskManager.ScheduleLoadFeedTask(feedId)
	}
}

func (c *FeedListController) HandleTaskScheduled() {
	c.appController.App.QueueUpdateDraw(func() {
		c.numUncompletedTasks++
		c.updateTaskStatusText()
	})
}

func (c *FeedListController) HandleTaskCompleted(r task.TaskResult) {
	c.appController.App.QueueUpdateDraw(func() {
		c.numUncompletedTasks--
		c.updateTaskStatusText()
		c.LoadFeedsFromStore()
	})
}

func (c *FeedListController) handleFeedSelected(idx int, text string, secondaryText string, shortcut rune) {
	feedId := c.listIdxToFeedId[idx]
	c.feedDetailController.SetDisplayedFeed(feedId)
	c.appController.SwitchToPage(pageFeedDetail)
}

func (c *FeedListController) updateTaskStatusText() {
	status := "All feeds updated"
	if c.numUncompletedTasks > 0 {
		status = fmt.Sprintf("Refreshing %v feed(s)...", c.numUncompletedTasks)
	}
	c.statusHeader.SetText(status)
}
