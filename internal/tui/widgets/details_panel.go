package widgets

import (
	"image"
	"sync"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/watzon/go-up/internal/types"
)

type DetailsPanel struct {
	Container *termui.Block
	URLView   *widgets.Paragraph
	Chart     *ResponseChart
	Stats     *StatsTable
	sync.Mutex
}

func NewDetailsPanel() *DetailsPanel {
	container := termui.NewBlock() // Changed from widgets.NewBlock() to termui.NewBlock()
	container.Border = true
	container.BorderStyle = termui.NewStyle(termui.ColorWhite)

	urlView := widgets.NewParagraph()
	urlView.TextStyle = termui.NewStyle(termui.ColorGreen)
	urlView.Border = false
	urlView.TextStyle.Modifier = termui.ModifierBold | termui.ModifierUnderline

	return &DetailsPanel{
		Container: container,
		URLView:   urlView,
		Chart:     NewResponseChart(),
		Stats:     NewStatsTable(),
	}
}

func (d *DetailsPanel) SetRect(x1, y1, x2, y2 int) {
	d.Lock()
	defer d.Unlock()

	d.Container.SetRect(x1, y1, x2, y2)

	// URL at the top
	urlStart := y1 + 1
	d.URLView.SetRect(x1+1, urlStart, x2-1, urlStart+1)

	// Stats at the bottom, 7 rows high
	statsHeight := 7
	statsStart := y2 - statsHeight - 1 // -1 for container border
	d.Stats.SetRect(x1+1, statsStart, x2-1, statsStart+statsHeight)

	// Chart fills remaining space between URL and stats
	chartStart := urlStart + 1
	chartHeight := statsStart - chartStart
	d.Chart.SetRect(x1+1, chartStart, x2-1, chartStart+chartHeight)

	// Initialize chart with stored stats if we have them
	d.Chart.InitializeWithCurrentWidth()
}

func (d *DetailsPanel) Update(status types.ServiceStatus) {
	d.Lock()
	defer d.Unlock()

	d.Container.Title = status.ServiceName
	d.URLView.Text = status.ServiceURL
	d.Chart.Update(status)
	d.Stats.Update(status)
}

func (d *DetailsPanel) Draw(buf *termui.Buffer) {
	d.Lock()
	defer d.Unlock()

	d.Container.Draw(buf)
	d.URLView.Draw(buf)
	d.Chart.Draw(buf)
	d.Stats.Draw(buf)
}

// GetRect implements termui.Drawable interface
func (d *DetailsPanel) GetRect() image.Rectangle {
	d.Lock()
	defer d.Unlock()
	return d.Container.GetRect()
}

// Lock is already implemented by embedding sync.Mutex
// Unlock is already implemented by embedding sync.Mutex

// Add this method to DetailsPanel
func (d *DetailsPanel) InitializeChart(stats []types.HistoricalStat, debug *DebugView) {
	d.Lock()
	defer d.Unlock()
	if debug != nil {
		debug.Printf("DetailsPanel initializing chart with %d stats", len(stats))
	}

	// Get terminal dimensions instead of container width
	termWidth, termHeight := termui.TerminalDimensions()
	if debug != nil {
		debug.Printf("Terminal dimensions: %dx%d", termWidth, termHeight)
	}

	// Calculate the width we should use (similar to App.resize logic)
	leftPanelWidth := termWidth / 4
	rightPanelWidth := termWidth - leftPanelWidth - 2 // -2 for padding

	if debug != nil {
		debug.Printf("Using right panel width: %d", rightPanelWidth)
	}

	if rightPanelWidth > 0 {
		d.Chart.InitializeHistory(stats, rightPanelWidth, debug)
	} else {
		if debug != nil {
			debug.Printf("Width calculation failed, storing stats for later")
		}
		d.Chart.StoreHistoricalStats(stats)
	}
}
