package tui

import (
    "fmt"
    "image"
    "github.com/gizak/termui/v3"
    "github.com/gizak/termui/v3/widgets"
    "github.com/watzon/go-up/internal/types"
)

type views struct {
    serviceList    *widgets.List
    rightPanel     *widgets.Paragraph
    urlView        *widgets.Paragraph
    statusDots     *termui.Canvas
    statsView      *widgets.Table
    debugView      *widgets.List
    selectedIndex  int
    statusHistory  map[string][]bool
    debugHistory   []string
    isPaused       bool
}

func newViews(services []string) *views {
    // Initialize service list
    serviceList := widgets.NewList()
    serviceList.Title = "Monitors"
    serviceList.TextStyle = termui.NewStyle(termui.ColorWhite)
    serviceList.WrapText = false

    // Initialize right panel with frame
    rightPanel := widgets.NewParagraph()
    rightPanel.Border = true
    rightPanel.BorderStyle = termui.NewStyle(termui.ColorWhite)
    rightPanel.Text = "" // The panel itself will be empty as it's just a container

    // Initialize URL view
    urlView := widgets.NewParagraph()
    urlView.TextStyle = termui.NewStyle(termui.ColorGreen)
    urlView.Border = false

    // Initialize status dots canvas
    statusDots := termui.NewCanvas()

    // Initialize stats view
    statsView := widgets.NewTable()
    statsView.Border = true
    statsView.TextStyle = termui.NewStyle(termui.ColorWhite)
    statsView.RowSeparator = true
    statsView.FillRow = false
    statsView.ColumnWidths = []int{15, 15, 15, 15, 15}
    statsView.Rows = [][]string{
        {"Response", "Avg. Response", "Uptime", "Uptime", "Cert Exp."},
        {"(current)", "(24-hour)", "(24-hour)", "(30-day)", ""},
        {"--", "--", "--", "--", "--"},
    }

    // Initialize debug view
    debugView := widgets.NewList()
    debugView.Title = "Debug Info (Press 'p' to pause)"  // Add pause hint to title
    debugView.TextStyle = termui.NewStyle(termui.ColorYellow)
    debugView.BorderStyle = termui.NewStyle(termui.ColorWhite)
    debugView.WrapText = false

    // Initialize status history with a smaller number of dots
    statusHistory := make(map[string][]bool)
    for _, service := range services {
        // Start with 30 dots instead of 60
        history := make([]bool, 30)
        // Initialize all to true (or false if you prefer)
        for i := range history {
            history[i] = true
        }
        statusHistory[service] = history
    }

    v := &views{
        serviceList:    serviceList,
        rightPanel:     rightPanel,
        urlView:        urlView,
        statusDots:     statusDots,
        statsView:      statsView,
        debugView:      debugView,
        selectedIndex:  0,
        statusHistory:  statusHistory,
        debugHistory:   make([]string, 0, 100),
        isPaused:       false,
    }

    v.resize()
    return v
}

func (v *views) resize() {
    termWidth, termHeight := termui.TerminalDimensions()
    
    leftPanelWidth := termWidth / 4
    rightPanelStart := leftPanelWidth + 1
    
    // Left panel
    v.serviceList.SetRect(0, 0, leftPanelWidth, termHeight)
    
    // Right panel with frame
    v.rightPanel.SetRect(rightPanelStart, 0, termWidth-1, termHeight-11) // Leave space for debug
    
    // URL view - now inside the right panel
    urlStart := 1 // Relative to panel start
    v.urlView.SetRect(rightPanelStart+1, urlStart+1, termWidth-2, urlStart+2)
    
    // Status dots - inside panel
    dotsStart := 3
    dotsHeight := 4
    v.statusDots.SetRect(rightPanelStart+1, dotsStart+1, termWidth-2, dotsStart+dotsHeight)
    
    // Stats table - inside panel
    statsStart := dotsStart + dotsHeight + 1
    statsHeight := 8
    v.statsView.SetRect(rightPanelStart+1, statsStart+1, termWidth-2, statsStart+statsHeight)

    // Debug view at bottom
    debugHeight := 10
    v.debugView.SetRect(rightPanelStart, termHeight-debugHeight, termWidth-1, termHeight)
}

func (v *views) debugf(format string, args ...interface{}) {
    // Add new message to debug history
    msg := fmt.Sprintf(format, args...)
    v.debugHistory = append(v.debugHistory, msg)

    // Keep last 100 messages
    if len(v.debugHistory) > 100 {
        v.debugHistory = v.debugHistory[len(v.debugHistory)-100:]
    }

    // Update the list rows
    v.debugView.Rows = v.debugHistory

    // Only auto-scroll if not paused and at bottom
    if !v.isPaused && len(v.debugHistory) > v.debugView.Inner.Dy() {
        v.debugView.ScrollBottom()
    }
}

func (v *views) togglePause() {
    v.isPaused = !v.isPaused
    if v.isPaused {
        v.debugView.Title = "Debug Info (PAUSED - Press 'p' to resume)"
    } else {
        v.debugView.Title = "Debug Info (Press 'p' to pause)"
        // Optionally scroll to bottom when unpausing
        v.debugView.ScrollBottom()
    }
}

func (v *views) handleDebugScroll(e termui.Event) {
    switch e.ID {
    case "<PageUp>":
        v.debugView.ScrollPageUp()
    case "<PageDown>":
        v.debugView.ScrollPageDown()
    case "<Home>":
        v.debugView.ScrollTop()
    case "<End>":
        v.debugView.ScrollBottom()
    }
}

func (v *views) updateStatusDots(serviceName string, isUp bool) {
    // Update history
    history := v.statusHistory[serviceName]
    for i := 0; i < len(history)-1; i++ {
        history[i] = history[i+1]
    }
    history[len(history)-1] = isUp
    v.statusHistory[serviceName] = history

    // Get canvas dimensions
    rect := v.statusDots.GetRect()
    v.statusDots = termui.NewCanvas()
    v.statusDots.SetRect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)

    // Debug canvas dimensions
    v.debugf("Canvas dimensions: (%d,%d) to (%d,%d)", 
        rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
    v.debugf("Canvas size: %d x %d", rect.Dx(), rect.Dy())

    // Calculate dot dimensions and spacing
    dotWidth := 2    // Width of each dot
    dotHeight := 2   // Height of each dot (make dots square)
    spacing := 1     // Space between dots

    // Calculate vertical position
    // Let's try to position dots in the middle of our canvas
    canvasHeight := rect.Dy()
    y := canvasHeight / 2
    v.debugf("Dot vertical position (y): %d (canvas height: %d)", y, canvasHeight)

    // Calculate horizontal layout
    availableWidth := rect.Dx() - 2  // Leave 1 pixel padding on each side
    maxDots := availableWidth / (dotWidth + spacing)
    v.debugf("Can fit %d dots in width %d (dot width: %d, spacing: %d)", 
        maxDots, availableWidth, dotWidth, spacing)

    // Get displayable history
    displayHistory := history
    if len(history) > maxDots {
        displayHistory = history[len(history)-maxDots:]
    }

    // Draw dots
    for i, status := range displayHistory {
        // Calculate x position
        x := 1 + (i * (dotWidth + spacing))  // Start with 1 pixel padding
        
        color := termui.ColorRed
        if status {
            color = termui.ColorGreen
        }

        v.debugf("Drawing dot %d at (%d,%d) status=%v", i, x, y, status)
        
        // Draw a square dot
        for dy := 0; dy < dotHeight; dy++ {
            for dx := 0; dx < dotWidth; dx++ {
                // The actual drawing point
                drawX := x + dx
                drawY := y + dy
                v.statusDots.SetPoint(image.Pt(drawX, drawY), color)
            }
        }
    }
}

func (v *views) updateDetailsView(status types.ServiceStatus) {
    // Update panel title
    v.rightPanel.Title = status.ServiceName

    // Update URL
    v.urlView.Text = status.ServiceURL
    v.debugf("URL: %s", status.ServiceURL)

    // Update status dots
    v.updateStatusDots(status.ServiceName, status.ResponseTime > 0)

    // Update stats
    v.statsView.Rows = [][]string{
        {"Response", "Avg. Response", "Uptime", "Uptime", "Cert Exp."},
        {"(current)", "(24-hour)", "(24-hour)", "(30-day)", ""},
        {
            fmt.Sprintf("%d ms", status.ResponseTime),
            fmt.Sprintf("%d", int(status.AvgResponseTime)),
            status.Uptime24Hours,
            status.Uptime30Days,
            status.CertificateExpiry,
        },
    }

    v.debugf("Stats: [%d ms %d %s %s %s]", 
        status.ResponseTime, 
        int(status.AvgResponseTime),
        status.Uptime24Hours,
        status.Uptime30Days,
        status.CertificateExpiry)
}

func (v *views) updateServiceList(services []string, currentStatus map[string]types.ServiceStatus) {
    rows := make([]string, len(services))
    for i, service := range services {
        status, exists := currentStatus[service]
        uptime := "Unknown"
        if exists {
            uptime = status.Uptime24Hours
        }
        statusIcon := "⚫" // Default gray
        if exists {
            if status.ResponseTime > 0 {
                statusIcon = "🟢"
            } else {
                statusIcon = "🔴"
            }
        }
        rows[i] = fmt.Sprintf("%s %s (%s)", statusIcon, service, uptime)
    }
    v.serviceList.Rows = rows
}

func (v *views) render() {
    termui.Render(
        v.serviceList,
        v.rightPanel,
        v.urlView,
        v.statusDots,
        v.statsView,
        v.debugView,
    )
}
