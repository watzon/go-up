package widgets

import (
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type HelpBar struct {
	*widgets.Paragraph
}

func NewHelpBar() *HelpBar {
	p := widgets.NewParagraph()
	p.Border = true
	p.BorderStyle = termui.NewStyle(termui.ColorWhite)
	p.TextStyle = termui.NewStyle(termui.ColorCyan)

	return &HelpBar{Paragraph: p}
}

func (h *HelpBar) UpdateHelp(hasDebug bool, isPaused bool) {
	baseHelp := "q: Quit | ↑/k: Up | ↓/j: Down"
	debugHelp := ""
	pauseHelp := ""

	if hasDebug {
		debugHelp = " | PgUp/PgDn: Scroll Debug | Home/End: Top/Bottom"
	}

	if isPaused {
		pauseHelp = " | p: Resume Monitor"
	} else {
		pauseHelp = " | p: Pause Monitor"
	}

	h.Text = baseHelp + pauseHelp + debugHelp
}
