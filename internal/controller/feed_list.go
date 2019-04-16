package controller

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/i18n"
	"github.com/wedaly/local-news/internal/store"
	"github.com/wedaly/local-news/internal/task"
	"sort"
	"strings"
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
	deleteConfirmController *DeleteConfirmController,
	feedStore *store.FeedStore,
	taskManager *task.TaskManager) *FeedListController {

	// Set up the list of feeds
	list := tview.NewList().
		ShowSecondaryText(false)
	list.Box.SetBorder(true).
		SetTitle(i18n.Gettext("All Feeds"))

	// Set up the header to display feed loading status
	statusHeader := tview.NewTextView()

	// Set up the footer to show help text
	// translators: the characters in parentheses are keyboard commands
	helpText := i18n.Gettext("(a) Add Feed   (r) Refresh All   (ESC) Quit")
	helpFooter := tview.NewTextView().
		SetText(helpText)

	// Set up a grid to hold the list, header, and footer
	grid := tview.NewGrid().
		SetRows(1, 0, 2).
		AddItem(statusHeader, 0, 0, 1, 1, 0, 0, false).
		AddItem(list, 1, 0, 1, 1, 0, 0, true).
		AddItem(helpFooter, 2, 0, 1, 1, 0, 0, false)

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
		nil,
		0,
	}
	list.SetSelectedFunc(c.handleFeedSelected)

	// Subscribe for task updates
	taskManager.Subscribe(c)

	// Subscribe for delete notifications
	deleteConfirmController.Subscribe(c)

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

func (c *FeedListController) HandleFeedDeleted(store.FeedId) {
	c.LoadFeedsFromStore()
}

func (c *FeedListController) LoadFeedsFromStore() {
	feedRecords, err := c.feedStore.RetrieveFeeds()
	if err != nil {
		panic(err)
	}

	// Sort the feeds ascending by name
	// (case-insensitive, locale-aware)
	sort.SliceStable(feedRecords, func(i, j int) bool {
		s1 := strings.ToLower(feedRecords[i].Name)
		s2 := strings.ToLower(feedRecords[j].Name)
		return i18n.CompareStrings(s1, s2)
	})

	// Look up the currently selected feed ID
	// so we can preserve the selection after reloading
	selectedIdx := c.list.GetCurrentItem()
	selectedFeedId := store.FeedId(-1)
	newSelectedIdx := -1

	if selectedIdx < len(c.listIdxToFeedId) {
		selectedFeedId = c.listIdxToFeedId[selectedIdx]
	}

	// Replace existing items with feeds from the database.
	// Keep track of the feed ID for each item in the list
	// so we can operate on them later.
	c.list.Clear()
	c.listIdxToFeedId = make([]store.FeedId, len(feedRecords))
	for i, feed := range feedRecords {
		c.list.AddItem(feed.Name, "", 0, nil)
		c.listIdxToFeedId[i] = feed.Id

		// Found the new idx for the previously selected feed
		if feed.Id == selectedFeedId {
			newSelectedIdx = i
		}
	}

	// Set the selected item back to the feed selected before the refresh
	// (unless it was deleted by the refresh, in which case keep the default)
	if newSelectedIdx >= 0 {
		c.list.SetCurrentItem(newSelectedIdx)
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
	status := i18n.Gettext("All feeds updated")
	if c.numUncompletedTasks > 0 {
		// translators: the argument is the number of feeds being refreshed
		refreshMsg := i18n.NGettext(
			"Refreshing %v feed...",
			"Refreshing %v feeds...",
			c.numUncompletedTasks)
		formattedCount := i18n.FormatNumber(c.numUncompletedTasks)
		status = fmt.Sprintf(refreshMsg, formattedCount)
	}
	c.statusHeader.SetText(status)
}
