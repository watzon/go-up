package widgets

import (
	"fmt"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type DebugView struct {
	*widgets.List
	history []string
}

func NewDebugView() *DebugView {
	list := widgets.NewList()
	list.Title = "Debug Info"
	list.TextStyle = termui.NewStyle(termui.ColorYellow)
	list.BorderStyle = termui.NewStyle(termui.ColorWhite)
	list.WrapText = false

	return &DebugView{
		List:    list,
		history: make([]string, 0, 100),
	}
}

func (d *DebugView) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	d.history = append(d.history, msg)

	// Keep last 100 messages
	if len(d.history) > 100 {
		d.history = d.history[len(d.history)-100:]
	}

	d.Rows = d.history
	d.ScrollBottom()
}

func (d *DebugView) HandleScroll(e termui.Event) {
	switch e.ID {
	case "<PageUp>":
		d.ScrollPageUp()
	case "<PageDown>":
		d.ScrollPageDown()
	case "<Home>":
		d.ScrollTop()
	case "<End>":
		d.ScrollBottom()
	}
}
