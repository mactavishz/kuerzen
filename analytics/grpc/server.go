package server

import (
	"context"
	"time"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
	store "github.com/mactavishz/kuerzen/store/analytics"
)

type AnalyticsServer struct {
	pb.UnimplementedAnalyticsServiceServer
	store store.AnalyticsStore
}

func NewAnalyticsServer(store store.AnalyticsStore) *AnalyticsServer {
	return &AnalyticsServer{
		store: store,
	}
}

func (s *AnalyticsServer) CreateShortURLEvent(ctx context.Context, req *pb.CreateShortURLEventRequest) (*pb.EventResponse, error) {
	event := &store.URLCreationEvent{
		ServiceName: req.ServiceName,
		URL:         req.Url,
		APIVer:      int(req.ApiVersion),
		Timestamp:   time.UnixMicro(req.Timestamp),
	}
	s.store.WriteURLCreationEvent(event)
	return &pb.EventResponse{Success: true}, nil
}

func (s *AnalyticsServer) RedirectShortURLEvent(ctx context.Context, req *pb.RedirectShortURLEventRequest) (*pb.EventResponse, error) {
	event := &store.URLRedirectEvent{
		ServiceName: req.ServiceName,
		ShortURL:    req.ShortUrl,
		LongURL:     req.LongUrl,
		APIVer:      int(req.ApiVersion),
		Timestamp:   time.UnixMicro(req.Timestamp),
	}
	s.store.WriteURLRedirectEvent(event)
	return &pb.EventResponse{Success: true}, nil
}
