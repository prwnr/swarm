package pkg

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var color = tcell.NewRGBColor(64, 69, 82)

func MakeViews(app *tview.Application) (*tview.List, *tview.List) {
	events := tview.NewList().ShowSecondaryText(true)
	events.SetBorder(true).SetBackgroundColor(color)
	events.SetSelectedBackgroundColor(color).SetSecondaryTextColor(tcell.ColorWhite)
	events.SetTitle("Active Streams list")

	messages := tview.NewList().ShowSecondaryText(false)
	messages.SetBorder(true).SetBackgroundColor(color)
	messages.SetSelectedBackgroundColor(color)
	messages.SetTitle("Stream messages list")

	app.SetFocus(events)

	return events, messages
}
