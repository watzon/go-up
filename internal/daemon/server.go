package daemon

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/watzon/go-up/internal/database"
)

func Start(host string, port int) {
	log.Printf("Starting daemon on %s:%d...", host, port)
	db, err := database.NewDB("go-up.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	log.Println("Initializing database...")
	if err := db.Init(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	log.Println("Creating new service...")
	service := NewService(db)
	err = rpc.Register(service)
	if err != nil {
		log.Fatalf("Error registering RPC service: %v", err)
	}
	log.Println("RPC service registered successfully")

	log.Println("Starting RPC server...")
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("Error starting RPC server: %v", err)
	}
	defer listener.Close()

	log.Printf("Daemon running and listening on %s:%d...", host, port)

	go service.
		periodicUpdate()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}
		log.Printf("New connection accepted from %v", conn.RemoteAddr())
		go rpc.ServeConn(conn)
	}
}
