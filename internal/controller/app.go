package controller

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/store"
	"github.com/wedaly/local-news/internal/task"
)

const (
	pageFeedList   = "feedList"
	pageAddFeed    = "addFeed"
	pageFeedDetail = "feedDetail"
)

// AppController controls the UI for the application,
// delegating to other controllers when necessary.
type AppController struct {
	App             *tview.Application
	pages           *tview.Pages
	pageControllers map[string]PageController
	currentPage     string
}

func NewAppController(
	feedStore *store.FeedStore,
	taskManager *task.TaskManager) *AppController {

	app := tview.NewApplication()
	pages := tview.NewPages()
	pageControllers := make(map[string]PageController, 0)
	ac := &AppController{app, pages, pageControllers, pageFeedList}
	app.SetInputCapture(ac.CaptureInput)

	// Set up the "feed details" page controller
	feedDetailController := NewFeedDetailController(ac, feedStore, taskManager)
	pageControllers[pageFeedDetail] = feedDetailController

	// Set up the "feed list" page controller
	feedListController := NewFeedListController(
		ac,
		feedDetailController,
		feedStore,
		taskManager)
	pageControllers[pageFeedList] = feedListController

	// Set up the "add feed" page controller.
	addFeedController := NewAddFeedController(ac, feedStore, taskManager)
	pageControllers[pageAddFeed] = addFeedController

	// Load initial data from database
	feedListController.LoadFeedsFromStore()

	// Add all pages to the tview `Pages` instance
	// This has to happen last so that the UI setup from the child controllers
	// is visible to tview.
	pages.AddPage(pageFeedList, feedListController.GetPage(), true, true)
	pages.AddPage(pageAddFeed, addFeedController.GetPage(), true, false)
	pages.AddPage(pageFeedDetail, feedDetailController.GetPage(), true, false)
	app.SetRoot(pages, true)

	return ac
}

// CaptureInput intercepts input to the application, delegating
// to child controllers as appropriate.  See the tview documentation for details.
func (c *AppController) CaptureInput(event *tcell.EventKey) *tcell.EventKey {
	// Don't let Ctrl-C exit the program, since we're using escape instead.
	// Ctrl-C is too close to Ctrl-V
	if event.Key() == tcell.KeyCtrlC {
		return nil
	}

	// Delegate to controller for current page
	pc := c.pageControllers[c.currentPage]
	return pc.HandleInput(event)
}

// SwitchToPage displays the specified page in the UI
// The page string must be a valid page name, as defined above.
func (c *AppController) SwitchToPage(page string) {
	c.App.QueueUpdateDraw(func() {
		c.pages.SwitchToPage(page)
		c.currentPage = page
	})
}
