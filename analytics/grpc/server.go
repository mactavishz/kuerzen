package grpc

import (
	"context"
	"time"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
	store "github.com/mactavishz/kuerzen/store/analytics"
)

type AnalyticsGRPCServer struct {
	pb.UnimplementedAnalyticsServiceServer
	store store.AnalyticsStore
}

func NewAnalyticsGRPCServer(store store.AnalyticsStore) *AnalyticsGRPCServer {
	return &AnalyticsGRPCServer{
		store: store,
	}
}

func (s *AnalyticsGRPCServer) CreateShortURLEvent(ctx context.Context, req *pb.CreateShortURLEventRequest) (*pb.EventResponse, error) {
	event := &store.URLCreationEvent{
		ServiceName: req.ServiceName,
		URL:         req.Url,
		APIVer:      req.ApiVersion,
		Timestamp:   time.UnixMicro(req.Timestamp),
	}
	s.store.WriteURLCreationEvent(event)
	return &pb.EventResponse{Success: true}, nil
}

func (s *AnalyticsGRPCServer) RedirectShortURLEvent(ctx context.Context, req *pb.RedirectShortURLEventRequest) (*pb.EventResponse, error) {
	event := &store.URLRedirectEvent{
		ServiceName: req.ServiceName,
		ShortURL:    req.ShortUrl,
		LongURL:     req.LongUrl,
		APIVer:      req.ApiVersion,
		Timestamp:   time.UnixMicro(req.Timestamp),
	}
	s.store.WriteURLRedirectEvent(event)
	return &pb.EventResponse{Success: true}, nil
}
