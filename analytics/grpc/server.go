package server

import (
	"context"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
)

type AnalyticsServer struct {
	pb.UnimplementedAnalyticsServiceServer
}

func NewAnalyticsServer() *AnalyticsServer {
	return &AnalyticsServer{}
}

func (s *AnalyticsServer) CreateShortURLEvent(ctx context.Context, req *pb.CreateShortURLEventRequest) (*pb.EventResponse, error) {
	// TODO: Implement the logic to create a short URL event
	return &pb.EventResponse{Success: true}, nil
}

func (s *AnalyticsServer) RedirectShortURLEvent(ctx context.Context, req *pb.RedirectShortURLEventRequest) (*pb.EventResponse, error) {
	// TODO: Implement the logic to redirect a short URL event
	return &pb.EventResponse{Success: true}, nil
}
