// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: pb/analytics.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AnalyticsService_CreateShortURLEvent_FullMethodName   = "/pb.AnalyticsService/CreateShortURLEvent"
	AnalyticsService_RedirectShortURLEvent_FullMethodName = "/pb.AnalyticsService/RedirectShortURLEvent"
)

// AnalyticsServiceClient is the client API for AnalyticsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AnalyticsServiceClient interface {
	// Record an event when a URL is created
	CreateShortURLEvent(ctx context.Context, in *CreateShortURLEventRequest, opts ...grpc.CallOption) (*EventResponse, error)
	// Record an event when a short URL is accessed
	RedirectShortURLEvent(ctx context.Context, in *RedirectShortURLEventRequest, opts ...grpc.CallOption) (*EventResponse, error)
}

type analyticsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAnalyticsServiceClient(cc grpc.ClientConnInterface) AnalyticsServiceClient {
	return &analyticsServiceClient{cc}
}

func (c *analyticsServiceClient) CreateShortURLEvent(ctx context.Context, in *CreateShortURLEventRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, AnalyticsService_CreateShortURLEvent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *analyticsServiceClient) RedirectShortURLEvent(ctx context.Context, in *RedirectShortURLEventRequest, opts ...grpc.CallOption) (*EventResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(EventResponse)
	err := c.cc.Invoke(ctx, AnalyticsService_RedirectShortURLEvent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AnalyticsServiceServer is the server API for AnalyticsService service.
// All implementations must embed UnimplementedAnalyticsServiceServer
// for forward compatibility.
type AnalyticsServiceServer interface {
	// Record an event when a URL is created
	CreateShortURLEvent(context.Context, *CreateShortURLEventRequest) (*EventResponse, error)
	// Record an event when a short URL is accessed
	RedirectShortURLEvent(context.Context, *RedirectShortURLEventRequest) (*EventResponse, error)
	mustEmbedUnimplementedAnalyticsServiceServer()
}

// UnimplementedAnalyticsServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAnalyticsServiceServer struct{}

func (UnimplementedAnalyticsServiceServer) CreateShortURLEvent(context.Context, *CreateShortURLEventRequest) (*EventResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShortURLEvent not implemented")
}
func (UnimplementedAnalyticsServiceServer) RedirectShortURLEvent(context.Context, *RedirectShortURLEventRequest) (*EventResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RedirectShortURLEvent not implemented")
}
func (UnimplementedAnalyticsServiceServer) mustEmbedUnimplementedAnalyticsServiceServer() {}
func (UnimplementedAnalyticsServiceServer) testEmbeddedByValue()                          {}

// UnsafeAnalyticsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AnalyticsServiceServer will
// result in compilation errors.
type UnsafeAnalyticsServiceServer interface {
	mustEmbedUnimplementedAnalyticsServiceServer()
}

func RegisterAnalyticsServiceServer(s grpc.ServiceRegistrar, srv AnalyticsServiceServer) {
	// If the following call pancis, it indicates UnimplementedAnalyticsServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AnalyticsService_ServiceDesc, srv)
}

func _AnalyticsService_CreateShortURLEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShortURLEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalyticsServiceServer).CreateShortURLEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AnalyticsService_CreateShortURLEvent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalyticsServiceServer).CreateShortURLEvent(ctx, req.(*CreateShortURLEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AnalyticsService_RedirectShortURLEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RedirectShortURLEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AnalyticsServiceServer).RedirectShortURLEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AnalyticsService_RedirectShortURLEvent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AnalyticsServiceServer).RedirectShortURLEvent(ctx, req.(*RedirectShortURLEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AnalyticsService_ServiceDesc is the grpc.ServiceDesc for AnalyticsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AnalyticsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.AnalyticsService",
	HandlerType: (*AnalyticsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateShortURLEvent",
			Handler:    _AnalyticsService_CreateShortURLEvent_Handler,
		},
		{
			MethodName: "RedirectShortURLEvent",
			Handler:    _AnalyticsService_RedirectShortURLEvent_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/analytics.proto",
}
