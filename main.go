package main

import (
	"fmt"
	"log"
	"stream-monitor/pkg"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/go-redis/redis"
)

var T *pkg.Terminal

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		Password:    "",
		DB:          0,
		ReadTimeout: -1,
	})

	T = pkg.NewTerminal()

	ui.Render(T.ToGrid())

	_, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect with Redis, err: %v", err))
	}

	go startMonitoring(client)

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Down>":
				T.List.ScrollDown()
			case "<Up>":
				T.List.ScrollUp()
			case "<C-d>":
				T.List.ScrollHalfPageDown()
			case "<C-u>":
				T.List.ScrollHalfPageUp()
			case "<C-f>":
				T.List.ScrollPageDown()
			case "<C-b>":
				T.List.ScrollPageUp()
			case "w":
				T.EventMessages.ScrollUp()
			case "s":
				T.EventMessages.ScrollDown()
			case "<Home>":
				T.List.ScrollTop()
			case "<End>":
				T.List.ScrollBottom()
			case "<Enter>":
				stream := getSelectedStream()
				if stream != nil {
					T.EventMessages.ScrollTop()
					T.EventMessages.Title = fmt.Sprintf("Event messages list: %s", stream.Name)
					T.EventMessages.Rows = stream.GetMessagesList()
				}
			case "r":
				stream, ID := getSelectedStream(), getSelectedMessageID()
				if stream != nil && ID != "" {
					message, err := stream.GetMessage(ID)
					if err != nil {
						break
					}

					T.Message.Title = fmt.Sprintf("Message %s", message.ID)
					T.Message.Text = message.ParseContent()
				}
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				T.Grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(T.ToGrid())
			}

			ui.Render(T.ToGrid())
		case <-ticker:
			ui.Render(T.ToGrid())
		}
	}
}

func getSelectedStream() *pkg.Stream {
	if len(T.List.Rows) == 0 {
		return nil
	}

	el := strings.Split(T.List.Rows[T.List.SelectedRow], " ")
	e := el[0]
	stream := T.StreamsList[e]
	return stream
}

func getSelectedMessageID() string {
	if len(T.EventMessages.Rows) == 0 {
		return ""
	}

	return T.EventMessages.Rows[T.EventMessages.SelectedRow]
}

func startMonitoring(c *redis.Client) {
	var events []string
	var errors int

	for {
		res, err := c.Do("MONITOR").String()
		if err != nil {
			fmt.Errorf(err.Error())
			errors++
			if errors > 5 {
				panic(fmt.Sprintf("MONITOR keeps failing, last error: %v", err))
			}

			continue
		}

		split := strings.Split(strings.Replace(res, "\"", "", -1), " ")
		if len(split) > 3 && strings.ToUpper(split[3]) == "XADD" {
			e := split[4]
			if sliceContains(events, e) {
				continue
			}

			events = append(events, e)
			stream := &pkg.Stream{Name: e}
			T.AddStream(stream)
			go readEvent(c, stream)
		}
	}
}

func readEvent(client *redis.Client, stream *pkg.Stream) {

	messages, _ := client.XRange(stream.Name, "-", "+").Result()
	for _, m := range messages {
		_ = m
		stream.MessagesCount++
		stream.AddMessage(m.ID, m.Values)
		ui.Render(T.ToGrid())
	}

	for {
		newMessages, _ := client.XRead(&redis.XReadArgs{
			Streams: []string{stream.Name, "$"},
			Block:   0,
		}).Result()
		for _, xStream := range newMessages {
			_ = xStream.Stream
			for _, m := range xStream.Messages {
				stream.MessagesCount++
				stream.AddMessage(m.ID, m.Values)
				ui.Render(T.ToGrid())
			}
		}
	}
}

// SliceContains checks if slice of strings contains given string
func sliceContains(s []string, l string) bool {
	for _, v := range s {
		if v == l {
			return true
		}
	}

	return false
}