package daemon

import (
    "log"
    "net"
    "net/rpc"
)

func Start() {
    service := NewService()
    rpc.Register(service)
    
    listener, err := net.Listen("tcp", ":1234")
    if err != nil {
        log.Fatalf("Error starting RPC server: %v", err)
    }
    defer listener.Close()
    
    log.Println("Daemon running and listening on port 1234...")
    
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
