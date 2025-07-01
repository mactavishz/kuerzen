package grpc

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
	store "github.com/mactavishz/kuerzen/store/analytics"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	maxElapsedTime       = 30 * time.Second      //Maximum total duration that all retries may last
	initialSleepInterval = 50 * time.Millisecond //The initial waiting time before the first retry
	maxSleepInterval     = 5 * time.Second       //The longest possible waiting time between two retries
	maxRetries           = 10                    //Upper limit for the number of retries if time is not the primary termination condition
)

var ErrMaxElapsedTimeExceeded = errors.New("max elapsed time for operation exceeded")
var ErrRetriesExhausted = errors.New("operation failed after all retries exhausted")

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

	startTime := time.Now()
	sleep := initialSleepInterval

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			ac.logger.Infof("SendURLCreationEvent operation cancelled: %v", ctx.Err())
			return ctx.Err()
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
					goto retryAttempt
				}
				ac.logger.Errorf("Failed to send URL creation event (non-retryable gRPC error %s): %v", st.Code().String(), err)
				return err
			}
			ac.logger.Infof("Attempt to send URL creation event failed (non-gRPC error), retrying: %v", err)
			goto retryAttempt
		}
		ac.logger.Infof("Successfully sent URL creation event to Analytics Service.")
		return nil

	retryAttempt:
		sleep = time.Duration(math.Min(float64(maxSleepInterval), float64(initialSleepInterval)+rand.Float64()*float64(3*sleep-initialSleepInterval)))
		ac.logger.Infof("Waiting for %v before next retry attempt for URL creation event (attempt %d)", sleep, i+1)

		if time.Since(startTime)+sleep >= maxElapsedTime {
			ac.logger.Errorf("SendURLCreationEvent operation timed out after %v", maxElapsedTime)
			return ErrMaxElapsedTimeExceeded
		}

		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			ac.logger.Errorf("SendURLCreationEvent operation cancelled during backoff: %v", ctx.Err())
			return ctx.Err()
		}
	}

	ac.logger.Errorf("Failed to send URL creation event after multiple retries")
	return ErrRetriesExhausted
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

	startTime := time.Now()
	sleep := initialSleepInterval

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			ac.logger.Infof("SendURLRedirectEvent operation cancelled: %v", ctx.Err())
			return ctx.Err()
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
					goto retryAttempt
				}
				ac.logger.Errorf("Failed to send URL redirect event (non-retryable gRPC error %s): %v", st.Code().String(), err)
				return err
			}
			ac.logger.Infof("Attempt to send URL redirect event failed (non-gRPC error), retrying: %v", err)
			goto retryAttempt
		}
		ac.logger.Infof("Successfully sent URL redirect event to Analytics Service.")
		return nil

	retryAttempt:
		sleep = time.Duration(math.Min(float64(maxSleepInterval), float64(initialSleepInterval)+rand.Float64()*float64(3*sleep-initialSleepInterval)))
		ac.logger.Infof("Waiting for %v before next retry attempt for URL redirect event (attempt %d)", sleep, i+1)

		if time.Since(startTime)+sleep >= maxElapsedTime {
			ac.logger.Errorf("SendURLRedirectEvent operation timed out after %v", maxElapsedTime)
			return ErrMaxElapsedTimeExceeded
		}

		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			ac.logger.Errorf("SendURLRedirectEvent operation cancelled during backoff: %v", ctx.Err())
			return ctx.Err()
		}
	}

	ac.logger.Errorf("Failed to send URL redirect event after multiple retries")
	return ErrRetriesExhausted
}

func (ac *AnalyticsGRPCClient) Close() error {
	return ac.conn.Close()
}
