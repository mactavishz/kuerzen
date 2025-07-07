package grpc

import (
	"context"
	"time"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
	"github.com/mactavishz/kuerzen/retries"
	store "github.com/mactavishz/kuerzen/store/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type AnalyticsGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.AnalyticsServiceClient
	logger *zap.SugaredLogger
}

func NewAnalyticsGRPCClient(addr string, logger *zap.SugaredLogger) (*AnalyticsGRPCClient, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 30 * time.Second,
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second, // send pings every 30 seconds if there is no activity
			Timeout:             60 * time.Second, // wait 60 seconds for ping responses
			PermitWithoutStream: true,             // allow pings even without active streams
		}),
	}

	// Connect to each client and send keys
	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &AnalyticsGRPCClient{
		conn:   conn,
		client: pb.NewAnalyticsServiceClient(conn),
		logger: logger,
	}, nil
}

func (ac *AnalyticsGRPCClient) SendURLCreationEvent(ctx context.Context, event *store.URLCreationEvent) func() retries.RetryableFuncObject {
	req := &pb.CreateShortURLEventRequest{
		ServiceName: event.ServiceName,
		Url:         event.URL,
		ApiVersion:  event.APIVer,
		Success:     event.Success,
		Timestamp:   event.Timestamp.UnixMicro(),
	}
	return func() retries.RetryableFuncObject {
		var rfo retries.RetryableFuncObject
		rfo.Ctx = ctx
		rfo.Logger = ac.logger
		select {
		case <-ctx.Done():
			ac.logger.Infof("SendURLCreationEvent operation cancelled: %v", ctx.Err())
			rfo.Err = ctx.Err()
			return rfo
		default:
		}
		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := ac.client.CreateShortURLEvent(grpcCtx, req)
		defer cancel()
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				if st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded || st.Code() == codes.Internal || st.Code() == codes.ResourceExhausted {
					ac.logger.Infof("Attempt to send URL creation event failed (%s), retrying: %v", st.Code().String(), err)
					rfo.Err = retries.ErrTransient
					return rfo
				}
				ac.logger.Errorf("Failed to send URL creation event (non-retryable gRPC error %s): %v", st.Code().String(), err)
				rfo.Err = err
				return rfo
			}
			ac.logger.Infof("Attempt to send URL creation event failed (non-gRPC error), retrying: %v", err)
			rfo.Err = retries.ErrTransient
			return rfo
		}
		ac.logger.Infof("Successfully sent URL creation event to Analytics Service.")
		rfo.Err = nil
		return rfo
	}
}

func (ac *AnalyticsGRPCClient) SendURLRedirectEvent(ctx context.Context, event *store.URLRedirectEvent) func() retries.RetryableFuncObject {
	req := &pb.RedirectShortURLEventRequest{
		ServiceName: event.ServiceName,
		ShortUrl:    event.ShortURL,
		LongUrl:     event.LongURL,
		ApiVersion:  event.APIVer,
		Success:     event.Success,
		Timestamp:   event.Timestamp.UnixMicro(),
	}
	return func() retries.RetryableFuncObject {
		var rfo retries.RetryableFuncObject
		rfo.Ctx = ctx
		rfo.Logger = ac.logger
		select {
		case <-ctx.Done():
			ac.logger.Infof("SendURLRedirectEvent operation cancelled: %v", ctx.Err())
			rfo.Err = ctx.Err()
			return rfo
		default:
		}
		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := ac.client.RedirectShortURLEvent(grpcCtx, req)
		defer cancel()
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				if st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded || st.Code() == codes.Internal || st.Code() == codes.ResourceExhausted {
					ac.logger.Infof("Attempt to send URL redirect event failed (%s), retrying: %v", st.Code().String(), err)
					rfo.Err = retries.ErrTransient
					return rfo
				}
				ac.logger.Errorf("Failed to send URL redirect event (non-retryable gRPC error %s): %v", st.Code().String(), err)
				rfo.Err = err
				return rfo
			}
			ac.logger.Infof("Attempt to send URL redirect event failed (non-gRPC error), retrying: %v", err)
			rfo.Err = retries.ErrTransient
			return rfo
		}
		ac.logger.Infof("Successfully sent URL redirect event to Analytics Service.")
		rfo.Err = nil
		return rfo
	}
}

func (ac *AnalyticsGRPCClient) Close() error {
	return ac.conn.Close()
}
