package controller

import (
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"net/url"
)

// AddFeedController handles the form for creating a new feed from a URL
type AddFeedController struct {
	appController *AppController
	form          *tview.Form
	urlField      *tview.InputField
}

func NewAddFeedController(appController *AppController) *AddFeedController {
	// Set up the form
	form := tview.NewForm().
		AddInputField("URL", "", 0, nil, nil).
		AddButton("OK", nil)
	form.SetBorder(true).SetTitle("Add feed")

	// Configure the URL input field
	urlField, ok := form.GetFormItem(0).(*tview.InputField)
	if !ok {
		panic("Could not retrieve input field from form")
	}
	urlField.SetPlaceholder("Press Ctrl-V to paste feed URL")
	urlField.SetPlaceholderTextColor(tcell.ColorBlack)

	okButton := form.GetButton(0)

	// Install event handlers for text changed and OK pressed
	c := &AddFeedController{appController, form, urlField}
	urlField.SetChangedFunc(c.handleUrlFieldChange)
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

	// TODO: kick off task to upsert the feed

	c.urlField.SetText("")
	c.appController.SwitchToPage(pageFeedList)
}

func (c *AddFeedController) showError() {
	c.appController.App.QueueUpdateDraw(func() {
		c.form.SetFieldBackgroundColor(tcell.ColorRed)
		c.form.SetFieldTextColor(tcell.ColorWhite)
	})
}

func (c *AddFeedController) hideError() {
	c.appController.App.QueueUpdateDraw(func() {
		c.form.SetFieldBackgroundColor(tview.Styles.ContrastBackgroundColor)
		c.form.SetFieldTextColor(tview.Styles.PrimaryTextColor)
	})
}

func validateUrl(s string) bool {
	url, err := url.ParseRequestURI(s)
	return err == nil && len(url.Host) > 0
}
