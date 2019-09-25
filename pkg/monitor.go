package pkg

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rivo/tview"
	"strings"
)

type Monitor struct {
	Redis   *redis.Client
	App     *tview.Application
	Events  *tview.List
	Streams *Streams
}

const StreamCommand = "XADD"

func (m *Monitor) StartMonitoring() {
	var events []string
	var errorsCount int

	for {
		res, err := m.Redis.Do("MONITOR").String()
		if err != nil {
			fmt.Errorf(err.Error())
			errorsCount++
			if errorsCount > 5 {
				panic(fmt.Sprintf("MONITOR keeps failing, last error: %v", err))
			}

			continue
		}

		split := strings.Split(strings.Replace(res, "\"", "", -1), " ")
		if len(split) > 3 && strings.ToUpper(split[3]) == StreamCommand {
			e := split[4]
			if sliceContains(events, e) {
				continue
			}

			events = append(events, e)
			stream := &Stream{Name: e}
			m.Events.AddItem(stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount), 0, nil)
			m.App.QueueUpdate(func() {})
			m.Streams.Push(stream)
			go m.readEvent(stream)
		}
	}
}

func (m *Monitor) readEvent(stream *Stream) {
	key := m.Events.FindItems(stream.Name, "", true, false)

	messages, _ := m.Redis.XRange(stream.Name, "-", "+").Result()
	for _, mes := range messages {
		_ = mes
		stream.MessagesCount++
		stream.AddMessage(mes.ID, mes.Values)
		m.App.QueueUpdateDraw(func() {
			m.Events.SetItemText(key[0], stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount))
		})
	}

	for {
		newMessages, _ := m.Redis.XRead(&redis.XReadArgs{
			Streams: []string{stream.Name, "$"},
			Block:   0,
		}).Result()
		for _, xStream := range newMessages {
			_ = xStream.Stream
			for _, mes := range xStream.Messages {
				stream.MessagesCount++
				stream.AddMessage(mes.ID, mes.Values)
				m.App.QueueUpdateDraw(func() {
					m.Events.SetItemText(key[0], stream.Name, fmt.Sprintf("- messages count: %d", stream.MessagesCount))
				})
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
