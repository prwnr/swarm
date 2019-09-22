package pkg

import (
	"errors"
	"fmt"
	"sort"
)

type Stream struct {
	Name          string
	MessagesCount int
	Messages      map[string]StreamMessage
}

func (s *Stream) AddMessage(id string, message map[string]interface{}) {
	if s.Messages == nil {
		s.Messages = make(map[string]StreamMessage)
	}

	s.Messages[id] = StreamMessage{
		ID:      id,
		Content: message,
	}
}

func (s *Stream) GetMessage(ID string) (*StreamMessage, error) {
	m, ok := s.Messages[ID]
	if !ok {
		return nil, errors.New("there are no messages with given ID")
	}

	return &m, nil
}

func (s *Stream) GetMessagesList() []string {
	var list []string
	for _, m :=range s.Messages {
		list = append(list, m.ID)
	}

	sort.Strings(list)

	return list
}

type StreamMessage struct {
	ID      string
	Content map[string]interface{}
}

func (m *StreamMessage) ParseContent() string {
	var content string
	for k, v := range m.Content {
		content += fmt.Sprintf("Field: %s\r\n", k)
		content += fmt.Sprintf("Value: %s\r\n\r\n", v)
	}

	return content
}
