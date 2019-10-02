package pkg

import (
	"errors"
	"fmt"
	"sort"
)

// Stream is a struct that holds messages of a stream and its name.
type Stream struct {
	// Name of the stream
	Name          string
	// Messages collection
	Messages      map[string]StreamMessage
}

// AddMessage to current stream by ID and message content.
func (s *Stream) AddMessage(id string, message map[string]interface{}) StreamMessage {
	if s.Messages == nil {
		s.Messages = make(map[string]StreamMessage)
	}

	streamMessage := StreamMessage{
		ID:      id,
		Content: message,
	}

	s.Messages[id] = streamMessage

	return streamMessage
}

// GetMessage from stream messages collection by ID.
func (s *Stream) GetMessage(ID string) (*StreamMessage, error) {
	m, ok := s.Messages[ID]
	if !ok {
		return nil, errors.New("there are no messages with given ID")
	}

	return &m, nil
}

// GetMessagesList returns array of messages IDs.
func (s *Stream) GetMessagesList() []string {
	var list []string
	for _, m := range s.Messages {
		list = append(list, m.ID)
	}

	sort.Strings(list)

	return list
}

// MessagesCount returns how many messages are in a stream.
func (s *Stream) MessagesCount() int {
	return len(s.Messages)
}

// StreamMessage with ID and Content
type StreamMessage struct {
	// ID of the message
	ID      string
	// Content of the message
	Content map[string]interface{}
}

// ParseContent transforms message content (which is a map[string]interface{})
// and returns it as a single string.
func (m *StreamMessage) ParseContent() string {
	var list []string
	for k, _ := range m.Content {
		list = append(list, k)
	}

	sort.Strings(list)

	var content string
	for _, i := range list {
		content += fmt.Sprintf("Field: %s\r\n", i)
		content += fmt.Sprintf("Value: %s\r\n\r\n", m.Content[i])
	}

	return content
}

// Streams holds collection of streams.
type Streams struct {
	Collection map[string]*Stream
}

// Push stream to collection.
func (s *Streams) Push(stream *Stream) {
	if s.Collection == nil {
		s.Collection = make(map[string]*Stream)
	}

	if _, ok := s.Collection[stream.Name]; ok {
		return
	}

	s.Collection[stream.Name] = stream
}

// Find returns stream by key (name of the stream).
func (s *Streams) Find(key string) *Stream {
	stream, ok := s.Collection[key]

	if !ok {
		return nil
	}

	return stream
}