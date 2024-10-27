package tui

import (
    "net/rpc"
    "github.com/watzon/go-up/internal/types"
)

type RPCClient struct {
    client *rpc.Client
}

func newRPCClient() (*RPCClient, error) {
    client, err := rpc.Dial("tcp", "localhost:1234")
    if err != nil {
        return nil, err
    }
    return &RPCClient{client: client}, nil
}

func (c *RPCClient) synchronizeMonitors(monitors []string) error {
    var response string
    return c.client.Call("Service.SynchronizeMonitors", monitors, &response)
}

func (c *RPCClient) getServiceStatus(serviceName string) (types.ServiceStatus, error) {
    var status types.ServiceStatus
    err := c.client.Call("Service.GetServiceStatus", serviceName, &status)
    return status, err
}

func (c *RPCClient) close() error {
    return c.client.Close()
}
