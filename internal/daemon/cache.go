package daemon

import (
    "sync"
    "github.com/watzon/go-up/internal/types"
)

// ServiceCache stores the latest status for each service
type ServiceCache struct {
    mu       sync.RWMutex
    services map[string]types.ServiceStatus
}

func newServiceCache() ServiceCache {
    return ServiceCache{
        services: make(map[string]types.ServiceStatus),
    }
}

func (c *ServiceCache) update(serviceName string, status types.ServiceStatus) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.services[serviceName] = status
}

func (c *ServiceCache) get(serviceName string) (types.ServiceStatus, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    status, exists := c.services[serviceName]
    return status, exists
}
