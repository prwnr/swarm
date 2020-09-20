package pkg

import (
	"github.com/go-redis/redis"
	"time"
)

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
	//var errorsCount int
	checkedKeys := make(map[string]string)
	tick := time.NewTicker(time.Second * 1).C
	for {
		select {
		case <-tick:
			keys, err := m.Redis.Keys("*").Result()
			if err != nil {
				LogError(err.Error())
				continue
			}

			for _, k := range keys {
				if _, ok := checkedKeys[k]; ok {
					continue
				}

				t, err := m.Redis.Type(k).Result()
				if err != nil {
					LogError(err.Error())
					continue
				}
				checkedKeys[k] = t
			}

			for k, t := range checkedKeys {
				found := m.Streams.Find(k)
				if found != nil || t != "stream" {
					continue
				}

				stream := &Stream{Name: k}
				m.Streams.Push(stream)
				m.emitStreamAdded(*stream)
				go m.readEvents(stream)
			}
		}
	}
}

func (m *Monitor) readEvents(stream *Stream) {
	messages, err := m.Redis.XRange(stream.Name, "-", "+").Result()
	if err != nil {
		LogWarning(err.Error())
	}

	for _, mes := range messages {
		_ = mes
		newMess := stream.AddMessage(mes.ID, mes.Values)
		m.emitMessageAdded(*stream, newMess)
	}

	for {
		newMessages, err := m.Redis.XRead(&redis.XReadArgs{
			Streams: []string{stream.Name, "$"},
			Block:   0,
		}).Result()

		if err != nil {
			LogWarning(err.Error())
		}

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
