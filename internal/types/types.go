package types

import (
	"time"
)

// ServiceStatus represents the status of a monitored service
type ServiceStatus struct {
	ServiceURL        string
	ServiceName       string
	ResponseTime      int
	AvgResponseTime   float64
	Uptime24Hours     float64
	Uptime30Days      float64
	CurrentStatus     bool
	CertificateExpiry time.Time
	IsActive          bool
}

type Monitor struct {
	ID       int
	Name     string
	URL      string
	IsActive bool
}

type HistoricalStat struct {
	ResponseTime int
	IsUp         bool
	Timestamp    time.Time
}
