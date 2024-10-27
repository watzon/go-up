package types

// ServiceStatus represents the status of a monitored service
type ServiceStatus struct {
    ServiceName       string
    ServiceURL       string
    ResponseTime      int
    AvgResponseTime   float64
    Uptime24Hours     string
    Uptime30Days      string
    CertificateExpiry string
    CurrentStatus     string
}
