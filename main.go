package main

import (
    "time"
    "github.com/gizak/termui/v3"
    "github.com/gizak/termui/v3/widgets"
    "github.com/spf13/cobra"
    "net/rpc"
    "net"
    "log"
    "fmt"
    "math/rand"
    "sync"
)

func main() {
    var rootCmd = &cobra.Command{
        Use: "go-up",
        Short: "Starts the TUI interface",
        Run: func(cmd *cobra.Command, args []string) {
            // Start the TUI
            log.Println("Starting TUI...")
            startTUI()
        },
    }

    var startDaemonCmd = &cobra.Command{
        Use:   "daemon",
        Short: "Starts the go-up daemon",
        Run: func(cmd *cobra.Command, args []string) {
            // Start the daemon
            log.Println("Starting daemon...")
            startDaemon()
        },
    }

    rootCmd.AddCommand(startDaemonCmd)
    rootCmd.Execute()
}

// Define the data structure for RPC communication
type ServiceStatus struct {
    ServiceName       string
    ResponseTime      int
    AvgResponseTime   float64
    Uptime24Hours     string
    Uptime30Days      string
    CertificateExpiry string
}

// Add a cache to store the latest status for each service
type ServiceCache struct {
    mu       sync.RWMutex
    services map[string]ServiceStatus
}

// Service represents the RPC service that will handle requests
type Service struct {
    Monitors map[string]ServiceStatus
    cache    ServiceCache
}

func NewService() *Service {
    return &Service{
        Monitors: make(map[string]ServiceStatus),
        cache: ServiceCache{
            services: make(map[string]ServiceStatus),
        },
    }
}

// SynchronizeMonitors processes the requested monitors, adding new ones and removing obsolete ones
func (s *Service) SynchronizeMonitors(requestedMonitors []string, reply *string) error {
    existingMonitors := make(map[string]bool)

    // Mark existing monitors
    for monitor := range s.Monitors {
        existingMonitors[monitor] = false
    }

    // Add new monitors or mark as existing
    for _, monitor := range requestedMonitors {
        if _, exists := s.Monitors[monitor]; !exists {
            s.Monitors[monitor] = ServiceStatus{ServiceName: monitor}
        }
        existingMonitors[monitor] = true
    }

    // Remove monitors that are no longer requested
    for monitor, isStillRequested := range existingMonitors {
        if !isStillRequested {
            delete(s.Monitors, monitor)
        }
    }

    *reply = "Monitors synchronized successfully"
    return nil
}

// Modified daemon to include periodic updates
func startDaemon() {
    service := NewService()
    rpc.Register(service)
    listener, err := net.Listen("tcp", ":1234")
    if err != nil {
        log.Fatalf("Error starting RPC server: %v", err)
    }
    defer listener.Close()
    log.Println("Daemon running and listening on port 1234...")

    // Start the update goroutine
    go service.periodicUpdate()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Connection error: %v", err)
            continue
        }
        go rpc.ServeConn(conn)
    }
}

// Periodic update function for the daemon
func (s *Service) periodicUpdate() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        s.cache.mu.Lock()
        for serviceName := range s.Monitors {
            s.cache.services[serviceName] = ServiceStatus{
                ServiceName:       serviceName,
                ResponseTime:      rand.Intn(500),
                AvgResponseTime:   rand.Float64() * 100,
                Uptime24Hours:     fmt.Sprintf("%.2f%%", rand.Float64()*100),
                Uptime30Days:      fmt.Sprintf("%.2f%%", rand.Float64()*100),
                CertificateExpiry: time.Now().AddDate(0, 0, rand.Intn(365)).Format("2006-01-02"),
            }
        }
        s.cache.mu.Unlock()
    }
}

// Modified GetServiceStatus to return cached data
func (s *Service) GetServiceStatus(serviceName string, status *ServiceStatus) error {
    s.cache.mu.RLock()
    defer s.cache.mu.RUnlock()
    
    if cachedStatus, exists := s.cache.services[serviceName]; exists {
        *status = cachedStatus
        return nil
    }
    return fmt.Errorf("service not found: %s", serviceName)
}

func startTUI() {
    requestedMonitors := []string{"Service A", "Service B", "Service C"}

    rpcClient, err := rpc.Dial("tcp", "localhost:1234")
    if err != nil {
        log.Fatalf("Error connecting to daemon: %v", err)
    }
    defer rpcClient.Close()

    var response string
    err = rpcClient.Call("Service.SynchronizeMonitors", requestedMonitors, &response)
    if err != nil {
        log.Fatalf("Error during handshake: %v", err)
    }

    if err := termui.Init(); err != nil {
        log.Fatalf("failed to initialize termui: %v", err)
    }
    defer termui.Close()

    serviceList := widgets.NewList()
    serviceList.Title = "Services"
    serviceList.Rows = requestedMonitors
    serviceList.TextStyle = termui.NewStyle(termui.ColorYellow)
    serviceList.WrapText = false
    serviceList.SetRect(0, 0, 30, 10)

    detailsTextView := widgets.NewParagraph()
    detailsTextView.Title = "Service Details"
    detailsTextView.Text = "Fetching service details..."
    detailsTextView.SetRect(31, 0, 90, 10)
    detailsTextView.BorderStyle.Fg = termui.ColorCyan

    selectedIndex := 0
    serviceList.SelectedRow = selectedIndex

    // Initial render
    termui.Render(serviceList, detailsTextView)

    // Start update goroutine
    updateChan := make(chan ServiceStatus)
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            var status ServiceStatus
            err := rpcClient.Call("Service.GetServiceStatus", requestedMonitors[selectedIndex], &status)
            if err != nil {
                log.Printf("Error calling RPC: %v", err)
                continue
            }
            updateChan <- status
        }
    }()

    uiEvents := termui.PollEvents()
    currentStatus := make(map[string]ServiceStatus)

    for {
        select {
        case status := <-updateChan:
            currentStatus[status.ServiceName] = status
            if status.ServiceName == requestedMonitors[selectedIndex] {
                updateDetailsView(detailsTextView, status)
                termui.Render(serviceList, detailsTextView)
            }

        case e := <-uiEvents:
            switch e.ID {
            case "q", "<C-c>":
                return
            case "j", "<Down>":
                if selectedIndex < len(serviceList.Rows)-1 {
                    selectedIndex++
                    serviceList.SelectedRow = selectedIndex
                }
            case "k", "<Up>":
                if selectedIndex > 0 {
                    selectedIndex--
                    serviceList.SelectedRow = selectedIndex
                }
            }

            // Use cached status when switching services
            if status, exists := currentStatus[requestedMonitors[selectedIndex]]; exists {
                updateDetailsView(detailsTextView, status)
            } else {
                detailsTextView.Text = "Fetching service details..."
            }
            termui.Render(serviceList, detailsTextView)
        }
    }
}

func updateDetailsView(view *widgets.Paragraph, status ServiceStatus) {
    view.Text = fmt.Sprintf(
        "Service Name: %s\n"+
            "Last Response Time: %d ms\n"+
            "Average Response Time (24h): %.2f ms\n"+
            "Uptime (24h): %s\n"+
            "Uptime (30d): %s\n"+
            "Certificate Expiration Date: %s",
        status.ServiceName, status.ResponseTime, status.AvgResponseTime,
        status.Uptime24Hours, status.Uptime30Days, status.CertificateExpiry)
}

func getServiceDetails(serviceName string) string {
    return "Fetching service details..."
}
