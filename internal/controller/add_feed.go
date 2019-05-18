package controller

import (
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/wedaly/local-news/internal/i18n"
	"github.com/wedaly/local-news/internal/store"
	"github.com/wedaly/local-news/internal/task"
	"net/url"
)

// AddFeedController handles the form for creating a new feed from a URL
type AddFeedController struct {
	appController *AppController
	config        i18n.Config
	feedStore     *store.FeedStore
	taskManager   *task.TaskManager
	form          *tview.Form
	urlField      *tview.InputField
}

func NewAddFeedController(
	appController *AppController,
	config i18n.Config,
	feedStore *store.FeedStore,
	taskManager *task.TaskManager) *AddFeedController {

	// Set up the form
	form := tview.NewForm().
		AddInputField(i18n.Gettext("URL"), "", 0, nil, nil).
		AddButton(i18n.Gettext("OK"), nil)
	form.SetBorder(true).SetTitle(
		i18n.Gettext("Add feed"))

	// Set initial colors based on localized config
	form.SetLabelColor(tcell.GetColor(config.FormLabelColor))
	form.SetButtonBackgroundColor(tcell.GetColor(config.FormButtonBackgroundColor))
	form.SetButtonTextColor(tcell.GetColor(config.FormButtonTextColor))
	form.SetFieldBackgroundColor(tcell.GetColor(config.FormFieldBackgroundColor))
	form.SetFieldTextColor(tcell.GetColor(config.FormFieldTextColor))

	// Configure the URL input field
	urlField, ok := form.GetFormItem(0).(*tview.InputField)
	if !ok {
		panic("Could not retrieve input field from form")
	}
	urlField.SetPlaceholder(
		i18n.Gettext("Press Ctrl-V to paste feed URL"))
	urlField.SetPlaceholderTextColor(tcell.ColorBlack)

	c := &AddFeedController{
		appController,
		config,
		feedStore,
		taskManager,
		form,
		urlField,
	}

	// Install event handlers for text changed and OK pressed
	urlField.SetChangedFunc(c.handleUrlFieldChange)
	okButton := form.GetButton(0)
	okButton.SetSelectedFunc(c.handleOkButton)

	return c
}

func (c *AddFeedController) GetPage() tview.Primitive {
	return c.form
}

func (c *AddFeedController) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	// Workaround for https://github.com/gdamore/tcell/issues/200
	// When pasting directly to the terminal, tcell truncates the input
	// to at most ten characters.
	// As a workaround, we copy the clipboard contents directly
	// to the input field when Ctrl-V is pressed.
	if event.Key() == tcell.KeyCtrlV {
		c.pasteClipboard()
		return nil
	}

	if event.Key() == tcell.KeyEscape {
		c.appController.SwitchToPage(pageFeedList)
		return nil
	}

	return event
}

func (c *AddFeedController) pasteClipboard() {
	clipboardText, err := clipboard.ReadAll()
	if err != nil {
		return
	}

	// This will replace the input field with the clipboard text.
	// That works okay for copying a full URL (the only input field in the app).
	// Unfortunately, `tview.InputField` doesn't expose a way to insert text
	// or manipulate the cursor position, so this is the best we can do
	// without creating a custom input field implementation.
	c.urlField.SetText(clipboardText)
}

func (c *AddFeedController) handleUrlFieldChange(text string) {
	if len(text) == 0 || validateUrl(text) {
		c.hideError()
	} else {
		c.showError()
	}
}

func (c *AddFeedController) handleOkButton() {
	urlText := c.urlField.GetText()
	if !validateUrl(urlText) {
		c.appController.App.SetFocus(c.urlField)
		return
	}

	// Create a placeholder database record for the feed
	feedId, err := c.feedStore.GetOrCreateFeedWithUrl(urlText)
	if err != nil {
		panic(err)
	}

	// Schedule background task to load the feed data
	c.taskManager.ScheduleLoadFeedTask(feedId)

	// Reset the UI
	c.urlField.SetText("")

	// Switch back to the feed list page
	c.appController.SwitchToPage(pageFeedList)
}

func (c *AddFeedController) showError() {
	c.appController.App.QueueUpdateDraw(func() {
		bg := tcell.GetColor(c.config.FormErrorBackgroundColor)
		txt := tcell.GetColor(c.config.FormErrorTextColor)
		c.form.SetFieldBackgroundColor(bg)
		c.form.SetFieldTextColor(txt)
	})
}

func (c *AddFeedController) hideError() {
	c.appController.App.QueueUpdateDraw(func() {
		bg := tcell.GetColor(c.config.FormFieldBackgroundColor)
		txt := tcell.GetColor(c.config.FormFieldTextColor)
		c.form.SetFieldBackgroundColor(bg)
		c.form.SetFieldTextColor(txt)
	})
}

func validateUrl(s string) bool {
	url, err := url.ParseRequestURI(s)
	return err == nil && len(url.Host) > 0
}
