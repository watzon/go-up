package tui

import (
	"net/rpc"
	"sync"
	"time"

	"github.com/watzon/go-up/internal/tui/widgets"
	"github.com/watzon/go-up/internal/types"
)

type RPCClient struct {
	addr   string
	client *rpc.Client
	mu     sync.Mutex
}

func newRPCClient(addr string) (*RPCClient, error) {
	c := &RPCClient{
		addr: addr,
	}
	err := c.connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *RPCClient) connect() error {
	client, err := rpc.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *RPCClient) ensureConnected() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return c.connect()
	}
	return nil
}

func (c *RPCClient) call(serviceMethod string, args interface{}, reply interface{}) error {
	err := c.ensureConnected()
	if err != nil {
		return err
	}

	err = c.client.Call(serviceMethod, args, reply)
	if err == rpc.ErrShutdown {
		c.client = nil
		time.Sleep(time.Second) // Wait a bit before reconnecting
		err = c.connect()
		if err != nil {
			return err
		}
		return c.client.Call(serviceMethod, args, reply)
	}
	return err
}

func (c *RPCClient) getServiceStatus(serviceName string) (types.ServiceStatus, error) {
	var status types.ServiceStatus
	err := c.call("Service.GetServiceStatus", serviceName, &status)
	return status, err
}

func (c *RPCClient) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

func (c *RPCClient) listMonitors() ([]types.Monitor, error) {
	var reply []types.Monitor
	err := c.call("Service.ListMonitors", struct{}{}, &reply)
	return reply, err
}

func (c *RPCClient) pauseMonitor(name string) error {
	var reply string
	return c.call("Service.PauseMonitor", name, &reply)
}

func (c *RPCClient) resumeMonitor(name string) error {
	var reply string
	return c.call("Service.ResumeMonitor", name, &reply)
}

// Update the getHistoricalStats method
func (c *RPCClient) getHistoricalStats(monitorID int, count int, debug *widgets.DebugView) ([]types.HistoricalStat, error) {
	var stats []types.HistoricalStat
	args := struct {
		MonitorID int
		Count     int
	}{
		MonitorID: monitorID,
		Count:     count,
	}
	err := c.call("Service.GetHistoricalStats", args, &stats)
	if err == nil && debug != nil {
		debug.Printf("Received %d historical stats for monitor %d", len(stats), monitorID)
	}
	return stats, err
}
