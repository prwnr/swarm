package pkg

import "fmt"

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

func (s *Stream) GetMessagesList() string {
	var list string
	for _, m :=range s.Messages {
		list += fmt.Sprintf("%s\r\n", m.ID)
	}

	return list
}

type StreamMessage struct {
	ID      string
	Content map[string]interface{}
}
