package widgets

import (
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/watzon/go-up/internal/types"
)

type ResponseChart struct {
	*widgets.BarChart
	data        []float64
	labels      []string
	statuses    []bool // Add this field to track up/down status
	storedStats []types.HistoricalStat
	debug       *DebugView
}

func NewResponseChart() *ResponseChart {
	chart := widgets.NewBarChart()
	chart.Border = false
	chart.BarWidth = 3
	chart.BarColors = []termui.Color{termui.ColorGreen}
	chart.LabelStyles = []termui.Style{termui.NewStyle(termui.ColorBlack)}
	chart.NumStyles = []termui.Style{termui.NewStyle(termui.ColorBlack)}

	return &ResponseChart{
		BarChart: chart,
		data:     make([]float64, 0),
		labels:   make([]string, 0),
		statuses: make([]bool, 0), // Initialize the statuses slice
	}
}

func (c *ResponseChart) Update(status types.ServiceStatus) {
	// Add new data point
	responseValue := float64(status.ResponseTime)
	if !status.CurrentStatus {
		responseValue = 0 // Use 0 for down status
	}

	c.data = append(c.data, responseValue)
	c.labels = append(c.labels, "")
	c.statuses = append(c.statuses, status.CurrentStatus)

	// Calculate max bars that can fit in current width
	maxBars := c.GetRect().Dx() / (c.BarWidth + 1)

	// Trim data to fit
	if len(c.data) > maxBars {
		c.data = c.data[len(c.data)-maxBars:]
		c.labels = c.labels[len(c.labels)-maxBars:]
		c.statuses = c.statuses[len(c.statuses)-maxBars:]
	}

	// Find max response time across all data points
	maxResponseTime := 0.0
	for i, v := range c.data {
		if c.statuses[i] && v > maxResponseTime {
			maxResponseTime = v
		}
	}

	// Update MaxVal based on max response time, clamped to 1000
	if maxResponseTime > 0 {
		c.MaxVal = maxResponseTime * 1.2 // Add 20% padding
		if c.MaxVal > 1000 {
			c.MaxVal = 1000
		}
		if c.debug != nil {
			c.debug.Printf("Update: Set MaxVal to %v (clamped to 1000)", c.MaxVal)
		}
	} else {
		c.MaxVal = 100 // Default when no response times
		if c.debug != nil {
			c.debug.Printf("Update: No valid response times found, setting MaxVal to 100")
		}
	}

	// Update colors based on status and clamp values to 999
	c.BarColors = make([]termui.Color, len(c.data))
	for i := range c.data {
		if !c.statuses[i] {
			c.BarColors[i] = termui.ColorRed
		} else if c.data[i] > 999 {
			c.data[i] = 999
			c.BarColors[i] = termui.ColorRed
		} else {
			c.BarColors[i] = termui.ColorGreen
		}
	}

	// Update chart data
	c.Data = c.data
	c.Labels = c.labels
}

// Add method to store stats for later initialization
func (c *ResponseChart) StoreHistoricalStats(stats []types.HistoricalStat) {
	c.storedStats = stats
}

// Add method to initialize with current width
func (c *ResponseChart) InitializeWithCurrentWidth() {
	if len(c.storedStats) > 0 {
		width := c.GetRect().Dx()
		if width > 0 {
			if c.debug != nil {
				c.debug.Printf("InitializeWithCurrentWidth: width is now %d, initializing stored stats", width)
			}
			c.InitializeHistory(c.storedStats, width, c.debug)
			c.storedStats = nil // Clear stored stats after initialization
		} else if c.debug != nil {
			c.debug.Printf("InitializeWithCurrentWidth: width still not valid (%d)", width)
		}
	}
}

// Add new method to calculate how many bars can fit
func (c *ResponseChart) CalculateMaxBars(containerWidth int) int {
	width := containerWidth - 2 // Subtract 2 for padding
	// Each bar takes BarWidth + 1 space (bar + gap)
	// We also need 1 space at the start for padding
	return (width - 1) / (c.BarWidth + 1)
}

// Update InitializeHistory to handle the data in the same way
func (c *ResponseChart) InitializeHistory(stats []types.HistoricalStat, containerWidth int, debug *DebugView) {
	c.debug = debug
	if debug != nil {
		debug.Printf("Initializing chart with %d stats", len(stats))
	}

	// Clear existing data
	c.data = make([]float64, 0, len(stats))
	c.labels = make([]string, 0, len(stats))
	c.statuses = make([]bool, 0, len(stats))

	// Stats already come in reverse chronological order (newest first)
	maxResponseTime := 0.0
	for _, stat := range stats {
		if debug != nil {
			debug.Printf("Processing stat: IsUp=%v, ResponseTime=%v",
				stat.IsUp, stat.ResponseTime)
		}

		responseValue := float64(stat.ResponseTime)
		if !stat.IsUp {
			responseValue = 0
		}

		if stat.IsUp && responseValue > maxResponseTime {
			maxResponseTime = responseValue
		}

		c.data = append(c.data, responseValue)
		c.labels = append(c.labels, "")
		c.statuses = append(c.statuses, stat.IsUp)
	}

	// Set chart properties
	if maxResponseTime > 0 {
		c.MaxVal = maxResponseTime * 1.2
		if debug != nil {
			debug.Printf("Set MaxVal to %v (120%% of max response time %v)", c.MaxVal, maxResponseTime)
		}
	} else {
		c.MaxVal = 100
		if debug != nil {
			debug.Printf("No valid response times found, setting MaxVal to 100")
		}
	}

	// Set colors based on status
	c.BarColors = make([]termui.Color, len(c.data))
	for i, isUp := range c.statuses {
		if isUp {
			c.BarColors[i] = termui.ColorGreen
		} else {
			c.BarColors[i] = termui.ColorRed
		}
	}

	// Set the data directly on the BarChart
	c.Data = c.data
	c.Labels = c.labels

	c.LabelStyles = []termui.Style{termui.NewStyle(termui.ColorBlack)}
	c.NumStyles = []termui.Style{termui.NewStyle(termui.ColorBlack)}

	if debug != nil {
		debug.Printf("Chart initialized with %d data points, MaxVal: %v", len(c.data), c.MaxVal)
		if len(c.data) > 0 {
			debug.Printf("First value: %v (status: %v), Last value: %v (status: %v)",
				c.data[0], c.statuses[0],
				c.data[len(c.data)-1], c.statuses[len(c.statuses)-1])
		}
	}
}
