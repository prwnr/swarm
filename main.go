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
		App:     app,
		Streams: &pkg.Streams{},
	}

	events := tview.NewList().ShowSecondaryText(true)
	events.SetBorder(true).SetBackgroundColor(color)
	events.SetSelectedBackgroundColor(color).SetSecondaryTextColor(tcell.ColorWhite)
	events.SetTitle("Active Streams list")

	monitor.Events = events

	messages := tview.NewList().ShowSecondaryText(false)
	messages.SetBorder(true).SetBackgroundColor(color)
	messages.SetSelectedBackgroundColor(color)
	messages.SetTitle("Stream messages list")

	app.SetFocus(events)

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

		app.QueueUpdateDraw(func() {})
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

	app.SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetFocus(events)
		}

		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}