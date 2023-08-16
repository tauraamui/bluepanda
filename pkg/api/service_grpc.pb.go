// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: service.proto

package api

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

// BluePandaClient is the client API for BluePanda service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BluePandaClient interface {
	Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResult, error)
}

type bluePandaClient struct {
	cc grpc.ClientConnInterface
}

func NewBluePandaClient(cc grpc.ClientConnInterface) BluePandaClient {
	return &bluePandaClient{cc}
}

func (c *bluePandaClient) Fetch(ctx context.Context, in *FetchRequest, opts ...grpc.CallOption) (*FetchResult, error) {
	out := new(FetchResult)
	err := c.cc.Invoke(ctx, "/bluepanda.BluePanda/Fetch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BluePandaServer is the server API for BluePanda service.
// All implementations must embed UnimplementedBluePandaServer
// for forward compatibility
type BluePandaServer interface {
	Fetch(context.Context, *FetchRequest) (*FetchResult, error)
	mustEmbedUnimplementedBluePandaServer()
}

// UnimplementedBluePandaServer must be embedded to have forward compatible implementations.
type UnimplementedBluePandaServer struct {
}

func (UnimplementedBluePandaServer) Fetch(context.Context, *FetchRequest) (*FetchResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Fetch not implemented")
}
func (UnimplementedBluePandaServer) mustEmbedUnimplementedBluePandaServer() {}

// UnsafeBluePandaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BluePandaServer will
// result in compilation errors.
type UnsafeBluePandaServer interface {
	mustEmbedUnimplementedBluePandaServer()
}

func RegisterBluePandaServer(s grpc.ServiceRegistrar, srv BluePandaServer) {
	s.RegisterService(&BluePanda_ServiceDesc, srv)
}

func _BluePanda_Fetch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FetchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BluePandaServer).Fetch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/bluepanda.BluePanda/Fetch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BluePandaServer).Fetch(ctx, req.(*FetchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BluePanda_ServiceDesc is the grpc.ServiceDesc for BluePanda service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BluePanda_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "bluepanda.BluePanda",
	HandlerType: (*BluePandaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Fetch",
			Handler:    _BluePanda_Fetch_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}
