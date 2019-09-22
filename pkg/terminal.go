package pkg

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sort"
)

type Terminal struct {
	StreamsList   map[string]*Stream
	EventMessages *widgets.List
	List          *widgets.List
	Message       *widgets.Paragraph
	Grid          *ui.Grid
}

func NewTerminal() *Terminal {
	eList := widgets.NewList()
	eList.Title = "List of events"
	eList.Rows = []string{""}
	eList.SelectedRowStyle.Fg = ui.ColorYellow

	mList := widgets.NewList()
	mList.Title = "Event messages list:"
	mList.Rows = []string{""}
	mList.SelectedRowStyle.Fg = ui.ColorBlue

	p := widgets.NewParagraph()
	p.Title = "Message"
	p.Text = ""

	g := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	g.SetRect(0, 0, termWidth, termHeight)

	g.Set(
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/2, eList),
			ui.NewCol(1.0/2, mList),
		),
		ui.NewRow(1.0/2,
			ui.NewCol(1.0, p),
		),
	)

	return &Terminal{
		EventMessages: mList,
		List:          eList,
		Grid:          g,
		Message:       p,
	}
}

func (t *Terminal) AddStream(s *Stream) {
	if t.StreamsList == nil {
		t.StreamsList = make(map[string]*Stream)
	}
	t.StreamsList[s.Name] = s
}

func (t *Terminal) ToGrid() *ui.Grid {
	t.List.Rows = []string{}

	var keys []string
	for k := range t.StreamsList {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, i := range keys {
		s := t.StreamsList[i]
		t.List.Rows = append(t.List.Rows, fmt.Sprintf("%s (%d)", s.Name, s.MessagesCount))
	}

	return t.Grid
}
