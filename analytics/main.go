package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	server "github.com/mactavishz/kuerzen/analytics/grpc"
	"github.com/mactavishz/kuerzen/analytics/pb"
	store "github.com/mactavishz/kuerzen/store/analytics"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	DEFAULT_GRPC_PORT    = "3003"
	DEFAULT_METRICS_PORT = "3002"
)

func main() {
	// We use sugar logger for better readability in development
	logger := zap.Must(zap.NewProduction()).Sugar()
	if os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "" {
		logger = zap.Must(zap.NewDevelopment()).Sugar()
	}
	defer logger.Sync()

	grpcPort := os.Getenv("ANALYTICS_GRPC_PORT")
	metricsPort := os.Getenv("ANALYTICS_PORT")

	if len(grpcPort) == 0 {
		grpcPort = DEFAULT_GRPC_PORT
	}

	if len(metricsPort) == 0 {
		metricsPort = DEFAULT_METRICS_PORT
	}

	// Setup prometheus metrics
	// Ref: https://github.com/grpc-ecosystem/go-grpc-middleware/tree/main/examples
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)

	client := influxdb2.NewClient(os.Getenv("ANALYTICS_DB_URL"), os.Getenv("DOCKER_INFLUXDB_INIT_ADMIN_TOKEN"))
	analyticsStore := store.NewInfluxDBAnalyticsStore(client, os.Getenv("DOCKER_INFLUXDB_INIT_ORG"), os.Getenv("DOCKER_INFLUXDB_INIT_BUCKET"))
	defer analyticsStore.Close()
	analyticsGRPCServer := server.NewAnalyticsGRPCServer(analyticsStore, logger)
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
		grpc.ChainUnaryInterceptor(srvMetrics.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(srvMetrics.StreamServerInterceptor()),
	}

	grpcServer := grpc.NewServer(serverOpts...)
	pb.RegisterAnalyticsServiceServer(grpcServer, analyticsGRPCServer)
	srvMetrics.InitializeMetrics(grpcServer)

	g := &run.Group{}
	g.Add(func() error {
		listener, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatalf("failed to listen: %v", err)
		}
		logger.Infof("Analytics service GRPC listening on port :%s", grpcPort)
		return grpcServer.Serve(listener)
	}, func(err error) {
		logger.Errorf("Failed to serve gRPC: %v", err)
		grpcServer.GracefulStop()
		grpcServer.Stop()
	})
	// Start metrics HTTP server
	httpServer := &http.Server{Addr: ":" + metricsPort}
	g.Add(func() error {
		http.Handle("/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte("{\"status\": \"healthy\"}"))
			if err != nil {
				logger.Errorf("failed to write health response: %v", err)
			}
		})
		logger.Infof("Metrics server listening on port :%s", metricsPort)
		return httpServer.ListenAndServe()
	}, func(err error) {
		if err := httpServer.Close(); err != nil {
			logger.Errorf("failed to serve metrics: %v", err)
		}
	})
	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))
	if err := g.Run(); err != nil {
		logger.Errorf("program interrupted: %v", err)
		os.Exit(1)
	}
}
