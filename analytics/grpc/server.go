package grpc

import (
	"context"
	"time"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
	store "github.com/mactavishz/kuerzen/store/analytics"
	"go.uber.org/zap"
)

type AnalyticsGRPCServer struct {
	pb.UnimplementedAnalyticsServiceServer
	store  store.AnalyticsStore
	logger *zap.SugaredLogger
}

func NewAnalyticsGRPCServer(store store.AnalyticsStore, logger *zap.SugaredLogger) *AnalyticsGRPCServer {
	return &AnalyticsGRPCServer{
		store:  store,
		logger: logger,
	}
}

func (s *AnalyticsGRPCServer) CreateShortURLEvent(ctx context.Context, req *pb.CreateShortURLEventRequest) (*pb.EventResponse, error) {
	event := &store.URLCreationEvent{
		ServiceName: req.ServiceName,
		URL:         req.Url,
		APIVer:      req.ApiVersion,
		Success:     req.Success,
		Timestamp:   time.UnixMicro(req.Timestamp),
	}
	s.store.WriteURLCreationEvent(event)
	s.logger.Infow("URL creation event recorded", "service", req.ServiceName)
	return &pb.EventResponse{Success: true}, nil
}

func (s *AnalyticsGRPCServer) RedirectShortURLEvent(ctx context.Context, req *pb.RedirectShortURLEventRequest) (*pb.EventResponse, error) {
	event := &store.URLRedirectEvent{
		ServiceName: req.ServiceName,
		ShortURL:    req.ShortUrl,
		LongURL:     req.LongUrl,
		APIVer:      req.ApiVersion,
		Success:     req.Success,
		Timestamp:   time.UnixMicro(req.Timestamp),
	}
	s.store.WriteURLRedirectEvent(event)
	s.logger.Infow("URL redirect event recorded", "service", req.ServiceName)
	return &pb.EventResponse{Success: true}, nil
}
