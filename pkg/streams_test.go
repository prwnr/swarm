package pkg

import (
	"reflect"
	"testing"
)

func TestStreams_Find(t *testing.T) {
	type fields struct {
		Collection map[string]*Stream
	}
	type args struct {
		key string
	}
	stream := &Stream{
		Name:     "Stream",
		Messages: nil,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Stream
	}{
		{
			"finds stream",
			fields{Collection: map[string]*Stream{"Stream": stream}},
			args{key: "Stream"},
			stream,
		},
		{
			"nil when there is no matchin stream",
			fields{Collection: map[string]*Stream{"Stream": stream}},
			args{key: "Not there"},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Streams{
				Collection: tt.fields.Collection,
			}
			if got := s.Find(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreams_Push(t *testing.T) {
	type fields struct {
		Collection map[string]*Stream
	}
	type args struct {
		stream *Stream
	}
	tests := []struct {
		name      string
		fields    fields
		key       string
		args      args
		wantCount int
	}{
		{"adds stream to collection",
			fields{Collection: nil},
			"Stream",
			args{stream: &Stream{
				Name:     "Stream",
				Messages: nil,
			}},
			1,
		},
		{"wont duplicate streams in collection when one already exists",
			fields{Collection: map[string]*Stream{
				"Stream": {
					Name:     "Stream",
					Messages: nil,
				},
			}},
			"Stream",
			args{stream: &Stream{
				Name:     "Stream",
				Messages: nil,
			}},
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Streams{
				Collection: tt.fields.Collection,
			}

			s.Push(tt.args.stream)
			if found := s.Find(tt.key); !reflect.DeepEqual(found, tt.args.stream) {
				t.Errorf("Find() = %v, want %v", found, tt.args.stream)
			}

			if got := len(s.Collection); got != tt.wantCount {
				t.Errorf("len(s.Collection) = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestStream_AddMessage(t *testing.T) {
	type args struct {
		id      string
		message map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want StreamMessage
	}{
		{"new message added", args{
			id:      "1",
			message: map[string]interface{}{"foo": "bar"},
		}, StreamMessage{
			ID:      "1",
			Content: map[string]interface{}{"foo": "bar"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stream{
				Name:     "Stream",
				Messages: nil,
			}
			if got := s.AddMessage(tt.args.id, tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_GetMessage(t *testing.T) {
	type fields struct {
		Name     string
		Messages map[string]StreamMessage
	}
	type args struct {
		ID string
	}
	message := &StreamMessage{
		ID:      "1",
		Content: map[string]interface{}{"foo": "bar"},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *StreamMessage
		wantErr bool
	}{
		{
			"finds message",
			fields{Name: "Stream", Messages: map[string]StreamMessage{"1": *message}},
			args{ID: "1"},
			message,
			false,
		},
		{
			"fails to find message",
			fields{Name: "Stream", Messages: map[string]StreamMessage{"1": *message}},
			args{ID: "2"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stream{
				Name:     tt.fields.Name,
				Messages: tt.fields.Messages,
			}
			got, err := s.GetMessage(tt.args.ID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMessage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_GetMessagesList(t *testing.T) {
	type fields struct {
		Name     string
		Messages map[string]StreamMessage
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{"single message on the list",
			fields{
				Name: "Stream",
				Messages: map[string]StreamMessage{
					"1": {
						ID:      "1",
						Content: map[string]interface{}{"foo": "bar"},
					},
				},
			},
			[]string{"1"},
		},
		{"multiple messages on the list",
			fields{
				Name: "Stream",
				Messages: map[string]StreamMessage{
					"1": {
						ID:      "1",
						Content: map[string]interface{}{"foo": "bar"},
					},
					"2": {
						ID:      "2",
						Content: map[string]interface{}{"foo": "bar"},
					},
				},
			},
			[]string{"1", "2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stream{
				Name:     tt.fields.Name,
				Messages: tt.fields.Messages,
			}
			if got := s.GetMessagesList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMessagesList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStream_MessagesCount(t *testing.T) {
	type fields struct {
		Name     string
		Messages map[string]StreamMessage
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"one message",
			fields{
				Name: "Stream",
				Messages: map[string]StreamMessage{
					"1": {
						ID:      "1",
						Content: map[string]interface{}{"foo": "bar"},
					},
				},
			},
			1,
		},
		{"mane messages",
			fields{
				Name: "Stream",
				Messages: map[string]StreamMessage{
					"1": {
						ID:      "1",
						Content: map[string]interface{}{"foo": "bar"},
					},
					"2": {
						ID:      "2",
						Content: map[string]interface{}{"foo": "bar"},
					},
					"3": {
						ID:      "3",
						Content: map[string]interface{}{"foo": "bar"},
					},
				},
			},
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stream{
				Name:     tt.fields.Name,
				Messages: tt.fields.Messages,
			}
			if got := s.MessagesCount(); got != tt.want {
				t.Errorf("MessagesCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamMessage_ParseContent(t *testing.T) {
	type fields struct {
		ID      string
		Content map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"simple content", fields{
			ID:      "1",
			Content: map[string]interface{}{"foo": "bar"},
		}, "Field: foo\r\nValue: bar\r\n\r\n"},
		{"complex content ordered", fields{
			ID:      "1",
			Content: map[string]interface{}{"b": "second", "a": "first", "c": "third"},
		}, "Field: a\r\nValue: first\r\n\r\nField: b\r\nValue: second\r\n\r\nField: c\r\nValue: third\r\n\r\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &StreamMessage{
				ID:      tt.fields.ID,
				Content: tt.fields.Content,
			}
			if got := m.ParseContent(); got != tt.want {
				t.Errorf("ParseContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
