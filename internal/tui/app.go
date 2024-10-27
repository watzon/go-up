package tui

import (
    "github.com/gizak/termui/v3"
    "log"
    "time"
    "github.com/watzon/go-up/internal/types"
)

func Start() {
    requestedMonitors := []string{"Service A", "Service B", "Service C"}

    client, err := newRPCClient()
    if err != nil {
        log.Fatalf("Error connecting to daemon: %v", err)
    }
    defer client.close()

    if err := client.synchronizeMonitors(requestedMonitors); err != nil {
        log.Fatalf("Error during handshake: %v", err)
    }

    if err := termui.Init(); err != nil {
        log.Fatalf("failed to initialize termui: %v", err)
    }
    defer termui.Close()

    v := newViews(requestedMonitors)
    lastStatus := make(map[string]types.ServiceStatus)

    // Initial render
    v.render()

    // Start update goroutine for all services
    updateChan := make(chan types.ServiceStatus)
    
    // Monitor each service independently
    for _, service := range requestedMonitors {
        go func(serviceName string) {
            ticker := time.NewTicker(5 * time.Second)
            defer ticker.Stop()

            // Initial status check
            status, err := client.getServiceStatus(serviceName)
            if err != nil {
                log.Printf("Initial status check error for %s: %v", serviceName, err)
            } else {
                updateChan <- status
            }

            // Periodic updates
            for range ticker.C {
                status, err := client.getServiceStatus(serviceName)
                if err != nil {
                    log.Printf("Error calling RPC for %s: %v", serviceName, err)
                    continue
                }
                updateChan <- status
            }
        }(service)
    }

    uiEvents := termui.PollEvents()

    for {
        select {
        case status := <-updateChan:
            // Only process updates if not paused
            if !v.isPaused {
                lastStatus[status.ServiceName] = status
                v.updateServiceList(requestedMonitors, lastStatus)
                
                if status.ServiceName == requestedMonitors[v.selectedIndex] {
                    v.updateDetailsView(status)
                }
                v.render()
            }

        case e := <-uiEvents:
            switch e.ID {
            case "q", "<C-c>":
                return
            case "p":
                v.togglePause()
            case "j", "<Down>":
                if v.selectedIndex < len(v.serviceList.Rows)-1 {
                    v.selectedIndex++
                    v.serviceList.SelectedRow = v.selectedIndex
                    if status, exists := lastStatus[requestedMonitors[v.selectedIndex]]; exists {
                        v.updateDetailsView(status)
                    }
                }
            case "k", "<Up>":
                if v.selectedIndex > 0 {
                    v.selectedIndex--
                    v.serviceList.SelectedRow = v.selectedIndex
                    if status, exists := lastStatus[requestedMonitors[v.selectedIndex]]; exists {
                        v.updateDetailsView(status)
                    }
                }
            case "<PageUp>", "<PageDown>", "<Home>", "<End>":
                v.handleDebugScroll(e)
            case "<Resize>":
                v.resize()
            }
            v.render()
        }
    }
}
