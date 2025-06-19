package grpc

import (
	"context"
	"time"

	pb "github.com/mactavishz/kuerzen/analytics/pb"
	store "github.com/mactavishz/kuerzen/store/analytics"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type AnalyticsGRPCClient struct {
	conn   *grpc.ClientConn
	client pb.AnalyticsServiceClient
}

func NewGRPCClient(addr string) (*AnalyticsGRPCClient, error) {
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
	_, err := ac.client.CreateShortURLEvent(ctx, req)
	if err != nil {
		return err
	}
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
	_, err := ac.client.RedirectShortURLEvent(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (ac *AnalyticsGRPCClient) Close() error {
	return ac.conn.Close()
}
