package tui

import (
	"fmt"
	"log"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/watzon/go-up/internal/tui/widgets"
	"github.com/watzon/go-up/internal/types"
)

type App struct {
	client           *RPCClient
	serviceList      *widgets.ServiceList
	details          *widgets.DetailsPanel
	debug            *widgets.DebugView
	help             *widgets.HelpBar
	monitors         []types.Monitor
	currentMonitorID int // Add this field
}

func NewApp(debugMode bool, daemonAddr string) (*App, error) {
	if err := termui.Init(); err != nil {
		return nil, err
	}

	client, err := newRPCClient(daemonAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}

	app := &App{
		client:           client,
		serviceList:      widgets.NewServiceList(),
		details:          widgets.NewDetailsPanel(),
		help:             widgets.NewHelpBar(),
		currentMonitorID: -1, // Initialize to invalid ID
	}

	if debugMode {
		app.debug = widgets.NewDebugView()
	}

	return app, nil
}

func (app *App) Run() error {
	defer termui.Close()
	defer app.Close() // Make sure we close the RPC client

	// Initial data fetch and setup
	if err := app.initialSetup(); err != nil {
		return err
	}

	// Initial render
	app.render()

	uiEvents := termui.PollEvents()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>", "<Escape>":
				termui.Clear()
				return nil
			case "j", "<Down>":
				if app.serviceList.GetSelectedIndex() < len(app.monitors) {
					app.serviceList.MoveSelection(1, len(app.monitors))
					app.updateSelectedMonitor(app.serviceList.GetSelectedIndex())
					app.render()
				}
			case "k", "<Up>":
				if app.serviceList.GetSelectedIndex() > 0 {
					app.serviceList.MoveSelection(-1, len(app.monitors))
					app.updateSelectedMonitor(app.serviceList.GetSelectedIndex())
					app.render()
				}
			case "<Resize>":
				payload := e.Payload.(termui.Resize)
				app.resize(payload.Width, payload.Height)
				app.render()
			case "p":
				app.handlePauseToggle()
				app.render()
			}
		case <-ticker.C:
			if err := app.refreshData(); err != nil {
				log.Printf("Error refreshing data: %v", err)
			}
			app.render()
		}
	}
}

func (app *App) resize(width, height int) int {
	leftPanelWidth := width / 4
	rightPanelStart := leftPanelWidth + 1
	rightPanelWidth := width - rightPanelStart - 1 // Subtract 1 for right border

	app.serviceList.SetRect(0, 0, leftPanelWidth, height-3)

	var chartWidth int
	if app.debug != nil {
		debugHeight := 10
		app.debug.SetRect(rightPanelStart, height-debugHeight-3, width-1, height-3)
		app.details.SetRect(rightPanelStart, 0, width-1, height-debugHeight-3)
		chartWidth = rightPanelWidth
	} else {
		app.details.SetRect(rightPanelStart, 0, width-1, height-3)
		chartWidth = rightPanelWidth
	}

	app.help.SetRect(0, height-3, width-1, height)

	return chartWidth
}

func (app *App) render() {
	termWidth, termHeight := termui.TerminalDimensions()
	chartWidth := app.resize(termWidth, termHeight)

	// Update help text
	selectedIdx := app.serviceList.GetSelectedIndex()
	isPaused := false
	if selectedIdx < len(app.monitors) {
		isPaused = app.serviceList.IsPaused(app.monitors[selectedIdx].Name)
	}
	app.help.UpdateHelp(app.debug != nil, isPaused)

	// Clear the terminal
	termui.Clear()

	// Draw the frame
	termui.Render(app.serviceList)
	termui.Render(app.details.Container) // Render container first
	termui.Render(app.details.URLView)   // Then child components
	termui.Render(app.details.Chart)
	termui.Render(app.details.Stats)
	termui.Render(app.help)

	if app.debug != nil {
		termui.Render(app.debug)
	}

	// After rendering, if we have a valid width, try to initialize any pending charts
	if chartWidth > 0 && selectedIdx < len(app.monitors) {
		monitor := app.monitors[selectedIdx]
		maxBars := app.details.Chart.CalculateMaxBars(chartWidth)

		if app.debug != nil {
			app.debug.Printf("Chart width: %d, Calculated maxBars: %d", chartWidth, maxBars)
		}

		if maxBars > 0 {
			// Get exactly the number of historical stats we need
			stats, err := app.client.getHistoricalStats(monitor.ID, maxBars, app.debug)
			if err != nil {
				if app.debug != nil {
					app.debug.Printf("Error fetching historical stats: %v", err)
				}
			} else {
				if len(stats) > 0 {
					app.details.InitializeChart(stats, app.debug)
				} else if app.debug != nil {
					app.debug.Printf("No historical stats received")
				}
			}
		}
	}
}

func (app *App) initialSetup() error {
	// Fetch initial monitors
	monitors, err := app.client.listMonitors()
	if err != nil {
		if app.debug != nil {
			app.debug.Printf("Error fetching monitors: %v", err)
		}
		return err
	}
	if app.debug != nil {
		app.debug.Printf("Found %d monitors", len(monitors))
	}
	app.monitors = monitors

	// Initialize status for all monitors
	currentStatus := make(map[string]types.ServiceStatus)
	for _, monitor := range monitors {
		status, err := app.client.getServiceStatus(monitor.Name)
		if err != nil {
			if app.debug != nil {
				app.debug.Printf("Error fetching initial status for %s: %v", monitor.Name, err)
			}
			continue
		}
		currentStatus[monitor.Name] = status
	}

	// Update service list with initial status
	app.serviceList.Update(monitors, currentStatus)

	// Initialize data for first monitor if available
	if len(monitors) > 0 {
		// Use the same function we use when changing selection
		app.updateSelectedMonitor(0)
	} else if app.debug != nil {
		app.debug.Printf("No monitors available for initialization")
	}

	// Set initial layout
	termWidth, termHeight := termui.TerminalDimensions()
	app.resize(termWidth, termHeight)

	// Do initial render
	app.render()

	return nil
}

func (app *App) handlePauseToggle() {
	selectedIdx := app.serviceList.GetSelectedIndex()
	if selectedIdx >= len(app.monitors) {
		return
	}

	monitor := app.monitors[selectedIdx]
	if app.serviceList.IsPaused(monitor.Name) {
		err := app.client.resumeMonitor(monitor.Name)
		if err != nil && app.debug != nil {
			app.debug.Printf("Error resuming monitor: %v", err)
		}
	} else {
		err := app.client.pauseMonitor(monitor.Name)
		if err != nil && app.debug != nil {
			app.debug.Printf("Error pausing monitor: %v", err)
		}
	}
	app.serviceList.TogglePause(monitor.Name)
}

func (app *App) refreshData() error {
	// Fetch latest monitors
	monitors, err := app.client.listMonitors()
	if err != nil {
		return err
	}
	app.monitors = monitors

	// Get current status for all monitors
	currentStatus := make(map[string]types.ServiceStatus)
	for _, monitor := range monitors {
		status, err := app.client.getServiceStatus(monitor.Name)
		if err != nil {
			if app.debug != nil {
				app.debug.Printf("Error fetching status for %s: %v", monitor.Name, err)
			}
			continue
		}
		currentStatus[monitor.Name] = status
	}

	// Update service list
	app.serviceList.Update(monitors, currentStatus)

	// Update details panel for selected service
	selectedIdx := app.serviceList.GetSelectedIndex()
	if selectedIdx < len(monitors) {
		if status, ok := currentStatus[monitors[selectedIdx].Name]; ok {
			app.details.Update(status)
		}
	}

	return nil
}

// Add this method to App struct
func (app *App) Close() error {
	if app.client != nil {
		return app.client.close()
	}
	return nil
}

// Update the updateSelectedMonitor method
func (app *App) updateSelectedMonitor(index int) {
	if index >= len(app.monitors) {
		return
	}

	monitor := app.monitors[index]

	// Only update if we're switching to a different monitor
	if app.currentMonitorID != monitor.ID {
		if app.debug != nil {
			app.debug.Printf("Updating selected monitor to %s (ID: %d)", monitor.Name, monitor.ID)
		}

		// Get current status
		status, err := app.client.getServiceStatus(monitor.Name)
		if err != nil {
			if app.debug != nil {
				app.debug.Printf("Error fetching status: %v", err)
			}
		} else {
			app.details.Update(status)
		}

		app.currentMonitorID = monitor.ID

		// Force a render which will handle chart initialization
		app.render()
	}
}
