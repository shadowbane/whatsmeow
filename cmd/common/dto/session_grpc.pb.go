// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: .protoc/session.proto

package dto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	SessionService_Connect_FullMethodName    = "/dto.SessionService/Connect"
	SessionService_Disconnect_FullMethodName = "/dto.SessionService/Disconnect"
	SessionService_Logout_FullMethodName     = "/dto.SessionService/Logout"
)

// SessionServiceClient is the client API for SessionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SessionServiceClient interface {
	Connect(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*SessionResponse, error)
	Disconnect(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*SessionResponse, error)
	Logout(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*SessionResponse, error)
}

type sessionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSessionServiceClient(cc grpc.ClientConnInterface) SessionServiceClient {
	return &sessionServiceClient{cc}
}

func (c *sessionServiceClient) Connect(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*SessionResponse, error) {
	out := new(SessionResponse)
	err := c.cc.Invoke(ctx, SessionService_Connect_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sessionServiceClient) Disconnect(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*SessionResponse, error) {
	out := new(SessionResponse)
	err := c.cc.Invoke(ctx, SessionService_Disconnect_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sessionServiceClient) Logout(ctx context.Context, in *SessionRequest, opts ...grpc.CallOption) (*SessionResponse, error) {
	out := new(SessionResponse)
	err := c.cc.Invoke(ctx, SessionService_Logout_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SessionServiceServer is the server API for SessionService service.
// All implementations must embed UnimplementedSessionServiceServer
// for forward compatibility
type SessionServiceServer interface {
	Connect(context.Context, *SessionRequest) (*SessionResponse, error)
	Disconnect(context.Context, *SessionRequest) (*SessionResponse, error)
	Logout(context.Context, *SessionRequest) (*SessionResponse, error)
	mustEmbedUnimplementedSessionServiceServer()
}

// UnimplementedSessionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSessionServiceServer struct {
}

func (UnimplementedSessionServiceServer) Connect(context.Context, *SessionRequest) (*SessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Connect not implemented")
}
func (UnimplementedSessionServiceServer) Disconnect(context.Context, *SessionRequest) (*SessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Disconnect not implemented")
}
func (UnimplementedSessionServiceServer) Logout(context.Context, *SessionRequest) (*SessionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedSessionServiceServer) mustEmbedUnimplementedSessionServiceServer() {}

// UnsafeSessionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SessionServiceServer will
// result in compilation errors.
type UnsafeSessionServiceServer interface {
	mustEmbedUnimplementedSessionServiceServer()
}

func RegisterSessionServiceServer(s grpc.ServiceRegistrar, srv SessionServiceServer) {
	s.RegisterService(&SessionService_ServiceDesc, srv)
}

func _SessionService_Connect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionServiceServer).Connect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SessionService_Connect_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionServiceServer).Connect(ctx, req.(*SessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SessionService_Disconnect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionServiceServer).Disconnect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SessionService_Disconnect_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionServiceServer).Disconnect(ctx, req.(*SessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SessionService_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SessionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionServiceServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SessionService_Logout_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionServiceServer).Logout(ctx, req.(*SessionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SessionService_ServiceDesc is the grpc.ServiceDesc for SessionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SessionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dto.SessionService",
	HandlerType: (*SessionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Connect",
			Handler:    _SessionService_Connect_Handler,
		},
		{
			MethodName: "Disconnect",
			Handler:    _SessionService_Disconnect_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _SessionService_Logout_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: ".protoc/session.proto",
}