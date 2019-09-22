package pkg

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"sort"
)

type Terminal struct {
	StreamsList   map[string]*Stream
	EventMessages *widgets.Paragraph
	List          *widgets.List
	Grid          *ui.Grid
}

func NewTerminal() *Terminal {
	l := widgets.NewList()
	l.Title = "List of events"
	l.Rows = []string{""}
	l.SelectedRowStyle.Fg = ui.ColorYellow

	p := widgets.NewParagraph()
	p.Title = "Event messages list: "
	p.Text = ""

	g := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	g.SetRect(0, 0, termWidth, termHeight)

	g.Set(
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/2, l),
			ui.NewCol(1.0/2, p),

		),
	)

	return &Terminal{
		EventMessages: p,
		List:          l,
		Grid:          g,
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
