package internal

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"swarm/pkg"
)

var color = tcell.NewRGBColor(64, 69, 82)

type Terminal struct {
	app                *tview.Application
	streams            *tview.List
	listeners          *tview.List
	listenersOutput    *tview.TextView
	messages           *tview.List
	messageContent     *tview.TextView
	activeStream       pkg.Stream
	printDefaultOutput chan bool
	Layout             *tview.Flex
}

func NewTerminal(app *tview.Application, withListener bool) *Terminal {
	t := &Terminal{
		app:                app,
		printDefaultOutput: make(chan bool),
	}

	tabs := makeTabs()
	pages := tview.NewPages()
	pages.AddPage("1", makeStreamsPage(t), true, true)
	pages.AddPage("2", makeListenersPage(t, withListener), true, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tabs, 1, 1, false).
		AddItem(pages, 0, 1, true)
	layout.SetBackgroundColor(color)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			if event.Rune() == 49 {
				tabs.Highlight("1").ScrollToHighlight()
				pages.SwitchToPage("1")
			} else if event.Rune() == 50 {
				tabs.Highlight("2").ScrollToHighlight()
				pages.SwitchToPage("2")
				if t.listeners != nil {
					t.printDefaultOutput <- true
				}
			}
		}

		return event
	})

	t.Layout = layout

	return t
}

// makeTabs creates the information about how many tabs/pages are there
// and what numbers are associated with them
func makeTabs() *tview.TextView {
	tabs := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
	tabs.SetBackgroundColor(color)
	tabs.Highlight("1")

	_, _ = fmt.Fprintf(tabs, `%d ["%d"][white]%s[white][""]  `, 1, 1, "Streams")
	_, _ = fmt.Fprintf(tabs, `%d ["%d"][white]%s[white][""]  `, 2, 2, "Listeners")

	return tabs
}

// makeStreamsPage prepares the content of the Streams page
// where it has active Events list, messages list of a selected Event
// and a content of a message
func makeStreamsPage(t *Terminal) *tview.Flex {
	events := tview.NewList().ShowSecondaryText(true)
	events.SetBorder(true).SetBackgroundColor(color)
	events.SetSelectedTextColor(color)
	events.SetSelectedBackgroundColor(tcell.ColorWhite)
	events.SetSecondaryTextColor(tcell.ColorWhite)
	events.SetTitle("Active Streams list")

	messages := tview.NewList().ShowSecondaryText(false)
	messages.SetBorder(true).SetBackgroundColor(color)
	messages.SetSelectedBackgroundColor(tcell.ColorWhite)
	messages.SetSelectedTextColor(color)
	messages.SetTitle("Stream messages list")
	messages.SetDoneFunc(func() {
		t.app.SetFocus(t.streams)
	})

	text := tview.NewTextView()
	text.SetBorder(true).SetTitle("Message content").SetBackgroundColor(color)
	text.SetChangedFunc(func() {
		t.app.QueueUpdateDraw(func() {})
	})
	text.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			t.app.SetFocus(t.streams)
		}
	})

	t.app.SetFocus(events)
	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(events, 0, 1, true).
			AddItem(messages, 0, 1, false), 0, 1, true).
		AddItem(text, 0, 3, false)
	flex.SetBackgroundColor(color)

	t.streams = events
	t.messages = messages
	t.messageContent = text

	return flex
}

// makeListenersPage prepares the content of the Listeners page
// where it shows active listeners and their statuses
// works only when artisan binary is properly detected
func makeListenersPage(t *Terminal, withListener bool) *tview.Flex {
	flex := tview.NewFlex()
	flex.SetBackgroundColor(color)
	if !withListener {
		text := tview.NewTextView()
		text.SetBorder(true).SetTitle("Info").SetBackgroundColor(color)
		text.SetChangedFunc(func() {
			t.app.QueueUpdateDraw(func() {})
		})
		_, _ = fmt.Fprint(text, "Artisan not detected. Listening is not available.")

		flex.AddItem(text, 0, 3, false)

		return flex
	}

	t.listeners = tview.NewList().ShowSecondaryText(true)
	t.listeners.SetBorder(true).SetBackgroundColor(color)
	t.listeners.SetSelectedBackgroundColor(tcell.ColorWhite)
	t.listeners.SetSelectedTextColor(color)
	t.listeners.SetSecondaryTextColor(tcell.ColorWhite)
	t.listeners.SetTitle("Listeners list")

	t.listenersOutput = tview.NewTextView()
	t.listenersOutput.SetBorder(true).SetTitle("Info").SetBackgroundColor(color)
	t.listenersOutput.SetChangedFunc(func() {
		t.app.QueueUpdateDraw(func() {})
	})
	t.listenersOutput.SetScrollable(true)

	t.listenersOutput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			t.app.SetFocus(t.listeners)
		}
	})

	flex.AddItem(t.listeners, 0, 1, true)
	flex.AddItem(t.listenersOutput, 0, 2, false)

	return flex
}

// BindMonitor binds terminal actions (view updates) to streamer monitor events.
func (t *Terminal) BindMonitor(monitor *pkg.Monitor) {
	monitor.OnNewStream(func(stream pkg.Stream) {
		t.streams.AddItem(stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount()), 0, nil)
	})

	monitor.OnNewMessage(func(stream pkg.Stream, message pkg.StreamMessage) {
		key := t.FindStreamKey(stream)
		if key < 0 {
			return
		}

		t.app.QueueUpdateDraw(func() {
			t.streams.SetItemText(key, stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount()))
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

// FindStreamKey returns match on a stream name from current streams list in terminal view.
func (t *Terminal) FindStreamKey(stream pkg.Stream) int {
	keys := t.streams.FindItems(stream.Name, "", true, false)

	var m string
	for _, k := range keys {
		m, _ = t.streams.GetItemText(k)
		if m == stream.Name {
			return k
		}
	}

	return -1
}

func (t *Terminal) BindListener(l *pkg.Listener) {
	go func() {
		for {
			select {
			case <-t.printDefaultOutput:
				if l.Items == nil {
					continue
				}

				main, _ := t.listeners.GetItemText(t.listeners.GetCurrentItem())
				lis, ok := l.Items[main]
				if !ok {
					continue
				}

				t.listenersOutput.Clear()
				_, _ = fmt.Fprint(t.listenersOutput, lis.ParseOutput())
			}
		}
	}()

	l.OnNewListener(func(listener pkg.StreamListener) {
		t.listeners.AddItem(listener.Name, fmt.Sprintf("Status: %s", listener.Status()), 0, nil)
	})

	l.OnListenerChange(func(listener pkg.StreamListener, lastOutput string) {
		key := t.FindListenerKey(listener.Name)
		if key < 0 {
			return
		}

		if key == t.listeners.GetCurrentItem() {
			_, _ = fmt.Fprint(t.listenersOutput, lastOutput)
		}

		t.app.QueueUpdateDraw(func() {
			t.listeners.SetItemText(key, listener.Name, fmt.Sprintf("Status: %s", listener.Status()))
		})
	})

	t.listeners.SetChangedFunc(func(key int, main string, secondary string, short rune) {
		lis, ok := l.Items[main]
		if !ok {
			return
		}

		t.listenersOutput.Clear()
		_, _ = fmt.Fprint(t.listenersOutput, lis.ParseOutput())
	})

	t.listeners.SetSelectedFunc(func(key int, main, secondary string, short rune) {
		t.app.SetFocus(t.listenersOutput)
	})
}

func (t *Terminal) FindListenerKey(name string) int {
	keys := t.listeners.FindItems(name, "", true, false)

	var m string
	for _, k := range keys {
		m, _ = t.streams.GetItemText(k)
		if m == name {
			return k
		}
	}

	return -1
}
