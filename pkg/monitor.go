package pkg

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
)

const StreamCommand = "XADD"

// Monitor connects with Redis and reads streams and messages from it
type Monitor struct {
	Redis           *redis.Client
	Streams         *Streams
	streamHandlers  []func(stream Stream)
	messageHandlers []func(stream Stream, message StreamMessage)
}

// NewMonitor creates monitor struct for usage.
func NewMonitor(c *redis.Client) *Monitor {
	return &Monitor{
		Redis:   c,
		Streams: &Streams{},
	}
}

// StartMonitoring uses Redis MONITOR command to catch all incoming streams
// and starts listening on them, adding them to Streams collection.
func (m *Monitor) StartMonitoring() {
	var events []string
	var errorsCount int

	for {
		res, err := m.Redis.Do("MONITOR").String()
		if err != nil {
			_ = fmt.Errorf(err.Error())
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
			m.Streams.Push(stream)
			m.emitStreamAdded(*stream)
			go m.readEvent(stream)
		}
	}
}

func (m *Monitor) readEvent(stream *Stream) {
	messages, _ := m.Redis.XRange(stream.Name, "-", "+").Result()
	for _, mes := range messages {
		_ = mes
		newMess := stream.AddMessage(mes.ID, mes.Values)
		m.emitMessageAdded(*stream, newMess)
	}

	for {
		newMessages, _ := m.Redis.XRead(&redis.XReadArgs{
			Streams: []string{stream.Name, "$"},
			Block:   0,
		}).Result()
		for _, xStream := range newMessages {
			_ = xStream.Stream
			for _, mes := range xStream.Messages {
				newMess := stream.AddMessage(mes.ID, mes.Values)
				m.emitMessageAdded(*stream, newMess)
			}
		}
	}
}

// OnNewStream assigns handlers that should be invoked when Monitor catches new stream by
func (m *Monitor) OnNewStream(handler func(stream Stream)) {
	m.streamHandlers = append(m.streamHandlers, handler)
}

// OnNewMessage assigns handlers that should be invoked when Monitor reads new message from a Stream
func (m *Monitor) OnNewMessage(handler func(stream Stream, message StreamMessage)) {
	m.messageHandlers = append(m.messageHandlers, handler)
}

func (m *Monitor) emitStreamAdded(stream Stream) {
	for _, l := range m.streamHandlers {
		l(stream)
	}
}

func (m *Monitor) emitMessageAdded(stream Stream, message StreamMessage) {
	for _, l := range m.messageHandlers {
		l(stream, message)
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
