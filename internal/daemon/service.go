package daemon

import (
    "fmt"
    "math/rand"
    "time"
    "strings"
    "github.com/watzon/go-up/internal/types"
)

// Service represents the RPC service that will handle requests
type Service struct {
    Monitors map[string]types.ServiceStatus
    cache    ServiceCache
}

func NewService() *Service {
    return &Service{
        Monitors: make(map[string]types.ServiceStatus),
        cache:    newServiceCache(),
    }
}

// SynchronizeMonitors processes the requested monitors
func (s *Service) SynchronizeMonitors(requestedMonitors []string, reply *string) error {
    existingMonitors := make(map[string]bool)

    for monitor := range s.Monitors {
        existingMonitors[monitor] = false
    }

    for _, monitor := range requestedMonitors {
        if _, exists := s.Monitors[monitor]; !exists {
            s.Monitors[monitor] = types.ServiceStatus{ServiceName: monitor}
        }
        existingMonitors[monitor] = true
    }

    for monitor, isStillRequested := range existingMonitors {
        if !isStillRequested {
            delete(s.Monitors, monitor)
        }
    }

    *reply = "Monitors synchronized successfully"
    return nil
}

func (s *Service) GetServiceStatus(serviceName string, status *types.ServiceStatus) error {
    // Add some randomization to make status changes more visible
    isUp := rand.Float64() > 0.2  // 80% chance of being up
    responseTime := 0
    if isUp {
        responseTime = rand.Intn(500) + 100  // Between 100-600ms when up
    }

    *status = types.ServiceStatus{
        ServiceName:       serviceName,
        ServiceURL:        fmt.Sprintf("https://%s.example.com", strings.ToLower(strings.ReplaceAll(serviceName, " ", "-"))),
        ResponseTime:      responseTime,
        AvgResponseTime:   float64(rand.Intn(100) + 50),
        Uptime24Hours:     fmt.Sprintf("%.2f%%", rand.Float64()*100),
        Uptime30Days:      fmt.Sprintf("%.2f%%", rand.Float64()*100),
        CertificateExpiry: time.Now().AddDate(0, 0, rand.Intn(365)).Format("2006-01-02"),
    }
    return nil
}

func (s *Service) periodicUpdate() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        for serviceName := range s.Monitors {
            status := types.ServiceStatus{
                ServiceName:       serviceName,
                ServiceURL:        fmt.Sprintf("https://%s.example.com", strings.ToLower(strings.ReplaceAll(serviceName, " ", "-"))),
                ResponseTime:      rand.Intn(500),
                AvgResponseTime:   float64(rand.Intn(100)),
                Uptime24Hours:     fmt.Sprintf("%.2f%%", rand.Float64()*100),
                Uptime30Days:      fmt.Sprintf("%.2f%%", rand.Float64()*100),
                CertificateExpiry: time.Now().AddDate(0, 0, rand.Intn(365)).Format("2006-01-02"),
            }
            s.cache.update(serviceName, status)
        }
    }
}
