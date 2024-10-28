package daemon

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/watzon/go-up/internal/database"
	"github.com/watzon/go-up/internal/types"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListMonitors(_ struct{}, reply *[]types.Monitor) error {
	monitors, err := s.db.ListMonitors()
	if err != nil {
		return err
	}
	*reply = monitors
	return nil
}

func (s *Service) AddMonitor(args struct{ Name, URL string }, reply *string) error {
	err := s.db.AddMonitor(args.Name, args.URL)
	if err != nil {
		*reply = fmt.Sprintf("Failed to add monitor %s for %s: %v", args.Name, args.URL, err)
		return err
	}

	responseTime, isUp, certExpiry := checkService(args.URL)
	if err := s.db.AddStats(args.Name, int(responseTime.Milliseconds()), isUp, certExpiry); err != nil {
		log.Printf("Warning: Failed to fetch initial stats for %s: %v", args.Name, err)
	}

	*reply = fmt.Sprintf("Monitor '%s' added for %s", args.Name, args.URL)
	return nil
}

func (s *Service) RemoveMonitor(name string, reply *string) error {
	err := s.db.RemoveMonitor(name)
	if err != nil {
		*reply = fmt.Sprintf("Failed to remove monitor %s: %v", name, err)
		return err
	}
	*reply = fmt.Sprintf("Monitor %s removed", name)
	return nil
}

func (s *Service) PauseMonitor(name string, reply *string) error {
	err := s.db.PauseMonitor(name)
	if err != nil {
		*reply = fmt.Sprintf("Failed to pause monitor %s: %v", name, err)
		return err
	}
	*reply = fmt.Sprintf("Monitor %s paused", name)
	return nil
}

func (s *Service) ResumeMonitor(name string, reply *string) error {
	err := s.db.ResumeMonitor(name)
	if err != nil {
		*reply = fmt.Sprintf("Failed to resume monitor %s: %v", name, err)
		return err
	}
	*reply = fmt.Sprintf("Monitor %s resumed", name)
	return nil
}

func (s *Service) GetServiceStatus(name string, reply *types.ServiceStatus) error {
	status, err := s.db.GetStats(name, 24*time.Hour)
	if err != nil {
		return err
	}
	*reply = status
	return nil
}

func (s *Service) GetHistoricalStats(args struct {
	MonitorID int
	Count     int
}, reply *[]types.HistoricalStat) error {
	stats, err := s.db.GetHistoricalStats(args.MonitorID, args.Count)
	if err != nil {
		return err
	}
	*reply = stats
	return nil
}

func (s *Service) periodicUpdate() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		monitors, err := s.db.ListMonitors()
		if err != nil {
			log.Printf("Error listing monitors: %v", err)
			continue
		}

		for _, monitor := range monitors {
			if !monitor.IsActive {
				continue
			}

			go func(m types.Monitor) {
				responseTime, isUp, certExpiry := checkService(m.URL)
				err := s.db.AddStats(m.Name, int(responseTime.Milliseconds()), isUp, certExpiry)
				if err != nil {
					log.Printf("Error adding stats for %s: %v", m.Name, err)
				}
			}(monitor)
		}
	}
}

func checkService(url string) (responseTime time.Duration, isUp bool, certExpiry time.Time) {
	start := time.Now()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(url)
	responseTime = time.Since(start)

	if err != nil {
		isUp = false
		return
	}
	defer resp.Body.Close()

	isUp = resp.StatusCode >= 200 && resp.StatusCode < 300

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		certExpiry = resp.TLS.PeerCertificates[0].NotAfter
	}

	return
}
