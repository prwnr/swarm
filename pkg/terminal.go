package pkg

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var color = tcell.NewRGBColor(64, 69, 82)

type Terminal struct {
	app            *tview.Application
	streams        *tview.List
	messages       *tview.List
	messageContent *tview.TextView
	activeStream   Stream
	Flex           *tview.Flex
}

func NewTerminal(app *tview.Application) *Terminal {
	events := tview.NewList().ShowSecondaryText(true)
	events.SetBorder(true).SetBackgroundColor(color)
	events.SetSelectedBackgroundColor(color).SetSecondaryTextColor(tcell.ColorWhite)
	events.SetTitle("Active Streams list")

	messages := tview.NewList().ShowSecondaryText(false)
	messages.SetBorder(true).SetBackgroundColor(color)
	messages.SetSelectedBackgroundColor(color)
	messages.SetTitle("Stream messages list")

	text := tview.NewTextView()
	text.SetBorder(true).SetTitle("Message content").SetBackgroundColor(color)
	text.SetChangedFunc(func() {
		app.QueueUpdateDraw(func() {})
	})

	app.SetFocus(events)

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(events, 0, 1, true).
			AddItem(messages, 0, 1, false), 0, 1, true).
		AddItem(text, 0, 3, false)

	flex.SetBackgroundColor(tcell.ColorDefault)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetFocus(events)
		}

		return event
	})

	t := &Terminal{
		app:            app,
		streams:        events,
		messages:       messages,
		messageContent: text,
		Flex:           flex,
	}

	return t
}

func (t *Terminal) BindMonitor(monitor *Monitor) {
	monitor.OnNewStream(func(stream Stream) {
		t.streams.AddItem(stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount()), 0, nil)
		t.app.QueueUpdateDraw(func() {})
	})

	monitor.OnNewMessage(func(stream Stream, message StreamMessage) {
		key := t.streams.FindItems(stream.Name, "", true, false)

		t.app.QueueUpdateDraw(func() {
			t.streams.SetItemText(key[0], stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount()))
		})

		if t.activeStream.Name == stream.Name && t.messages.GetFocusable().HasFocus() {
			t.messages.AddItem(message.ID, stream.Name, 0, nil)
		}
	})

	t.streams.SetSelectedFunc(func(key int, main, secondary string, short rune) {
		s := monitor.Streams.Find(main)
		if s == nil {
			return
		}

		t.messages.SetTitle(main)
		t.messages.Clear()
		for _, m := range s.GetMessagesList() {
			t.messages.AddItem(m, s.Name, 0, nil)
		}

		t.app.QueueUpdate(func() {})
		t.app.SetFocus(t.messages)
		t.activeStream = *s
	})

	t.messages.SetSelectedFunc(func(key int, main, secondary string, short rune) {
		s := monitor.Streams.Find(secondary)
		if s == nil {
			return
		}

		m, err := s.GetMessage(main)
		if err != nil {
			return
		}

		t.messageContent.Clear()
		_, _ = fmt.Fprint(t.messageContent, m.ParseContent())
	})
}
