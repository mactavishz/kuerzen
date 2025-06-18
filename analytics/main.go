package main

import (
	"log"
	"net"
	"os"
	"time"

	server "github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/analytics/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const DEFAULT_PORT = "3002"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	port := os.Getenv("ANALYTICS_PORT")

	if len(port) == 0 {
		port = DEFAULT_PORT
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("failed to listen: %v", err)
		os.Exit(1)
	}
	log.Printf("Analytics service listening on port :%s", port)

	gprcServer := server.NewAnalyticsServer()
	// Define keepalive server parameters
	kasp := keepalive.ServerParameters{
		Time:    30 * time.Second, // Ping the client if it is idle for 30 seconds to ensure the connection is still active
		Timeout: 60 * time.Second, // Wait 60 second for the ping ack before assuming the connection is dead
	}

	// Define keepalive enforcement policy
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}
	serverOpts := []grpc.ServerOption{
		grpc.KeepaliveParams(kasp),
		grpc.KeepaliveEnforcementPolicy(kaep),
	}

	gServer := grpc.NewServer(serverOpts...)
	pb.RegisterAnalyticsServiceServer(gServer, gprcServer)
	if err := gServer.Serve(listener); err != nil {
		log.Printf("failed to serve: %v", err)
		os.Exit(1)
	}
}
