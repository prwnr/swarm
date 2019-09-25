package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/go-redis/redis"
	"github.com/rivo/tview"
	"stream-monitor/pkg"
)

var app *tview.Application
var color = tcell.NewRGBColor(64, 69, 82)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		Password:    "",
		DB:          0,
		ReadTimeout: -1,
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect with Redis, err: %v", err))
	}

	app = tview.NewApplication()
	monitor := &pkg.Monitor{
		Redis:   client,
		Streams: &pkg.Streams{},
	}

	events, messages := pkg.MakeViews(app)

	monitor.OnNewStream(func(stream pkg.Stream) {
		events.AddItem(stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount), 0, nil)
		app.QueueUpdate(func() {})
	})

	monitor.OnNewMessage(func(stream pkg.Stream, message pkg.StreamMessage) {
		key := events.FindItems(stream.Name, "", true, false)

		app.QueueUpdateDraw(func() {
			events.SetItemText(key[0], stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount))
		})
	})

	go monitor.StartMonitoring()

	box := tview.NewTextView()
	box.SetBorder(true).SetTitle("Message content").SetBackgroundColor(color)
	box.SetChangedFunc(func() {
		app.Draw()
	})

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(events, 0, 1, true).
			AddItem(messages, 0, 1, false), 0, 1, true).
		AddItem(box, 0, 3, false)

	flex.SetBackgroundColor(tcell.ColorDefault)

	events.SetSelectedFunc(func(key int, main, secondary string, short rune) {
		s := monitor.Streams.Find(main)
		if s == nil {
			return
		}

		messages.SetTitle(main)
		messages.Clear()
		for _, m := range s.GetMessagesList() {
			messages.AddItem(m, s.Name, 0, nil)
		}

		app.QueueUpdate(func() {})
		app.SetFocus(messages)
	})

	messages.SetSelectedFunc(func(key int, main, secondary string, short rune) {
		s := monitor.Streams.Find(secondary)
		if s == nil {
			return
		}

		m, err := s.GetMessage(main)
		if err != nil {
			return
		}

		box.Clear()
		_, _ = fmt.Fprint(box, m.ParseContent())
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetFocus(events)
		}

		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
