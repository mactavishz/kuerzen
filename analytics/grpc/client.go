package grpc

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	pb "github.com/mactavishz/kuerzen/analytics/pb"
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

func (ac *AnalyticsGRPCClient) SendURLCreationEvent(ctx context.Context, event *store.URLCreationEvent) error {
	req := &pb.CreateShortURLEventRequest{
		ServiceName: event.ServiceName,
		Url:         event.URL,
		ApiVersion:  event.APIVer,
		Success:     event.Success,
		Timestamp:   event.Timestamp.UnixMicro(),
	}

	operation := func() error {
		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := ac.client.CreateShortURLEvent(grpcCtx, req)
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				if st.Code() == codes.Unavailable ||
					st.Code() == codes.DeadlineExceeded ||
					st.Code() == codes.Internal ||
					st.Code() == codes.ResourceExhausted {
					ac.logger.Warnf("Attempt to send URL creation event failed (%s), retrying: %v", st.Code().String(), err)
					return err
				}
				ac.logger.Errorf("Failed to send URL creation event (non-retryable code %s): %v", st.Code().String(), err)
				return backoff.Permanent(err)
			}
			ac.logger.Warnf("Attempt to send URL creation event failed (non-gRPC error), retrying: %v", err)
			return err
		}
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 15 * time.Second
	b.InitialInterval = 50 * time.Millisecond
	b.MaxInterval = 2 * time.Second
	b.RandomizationFactor = 0.5

	err := backoff.Retry(operation, backoff.WithContext(b, ctx))

	if err != nil {
		ac.logger.Errorf("Failed to send URL creation event after multiple retries: %v", err)
		return err
	}

	ac.logger.Debugf("Successfully sent URL creation event to Analytics Service.")
	return nil
}

func (ac *AnalyticsGRPCClient) SendURLRedirectEvent(ctx context.Context, event *store.URLRedirectEvent) error {
	req := &pb.RedirectShortURLEventRequest{
		ServiceName: event.ServiceName,
		ShortUrl:    event.ShortURL,
		LongUrl:     event.LongURL,
		ApiVersion:  event.APIVer,
		Success:     event.Success,
		Timestamp:   event.Timestamp.UnixMicro(),
	}

	operation := func() error {
		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := ac.client.RedirectShortURLEvent(grpcCtx, req)
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				if st.Code() == codes.Unavailable ||
					st.Code() == codes.DeadlineExceeded ||
					st.Code() == codes.Internal ||
					st.Code() == codes.ResourceExhausted {
					ac.logger.Warnf("Attempt to send URL creation event failed (%s), retrying: %v", st.Code().String(), err)
					return err
				}
				ac.logger.Errorf("Failed to send URL creation event (non-retryable code %s): %v", st.Code().String(), err)
				return backoff.Permanent(err)
			}
			ac.logger.Warnf("Attempt to send URL creation event failed (non-gRPC error), retrying: %v", err)
			return err
		}
		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 15 * time.Second
	b.InitialInterval = 50 * time.Millisecond
	b.MaxInterval = 2 * time.Second
	b.RandomizationFactor = 0.5

	err := backoff.Retry(operation, backoff.WithContext(b, ctx))

	if err != nil {
		ac.logger.Errorf("Failed to send URL creation event after multiple retries: %v", err)
		return err
	}

	ac.logger.Debugf("Successfully sent URL creation event to Analytics Service.")
	return nil
}

func (ac *AnalyticsGRPCClient) Close() error {
	return ac.conn.Close()
}
