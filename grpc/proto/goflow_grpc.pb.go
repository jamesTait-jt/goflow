// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: grpc/proto/goflow.proto

package goflow

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
	GoFlow_PushTask_FullMethodName  = "/goflow.GoFlow/PushTask"
	GoFlow_GetResult_FullMethodName = "/goflow.GoFlow/GetResult"
)

// GoFlowClient is the client API for GoFlow service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GoFlowClient interface {
	PushTask(ctx context.Context, in *PushTaskRequest, opts ...grpc.CallOption) (*PushTaskReply, error)
	GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*GetResultReply, error)
}

type goFlowClient struct {
	cc grpc.ClientConnInterface
}

func NewGoFlowClient(cc grpc.ClientConnInterface) GoFlowClient {
	return &goFlowClient{cc}
}

func (c *goFlowClient) PushTask(ctx context.Context, in *PushTaskRequest, opts ...grpc.CallOption) (*PushTaskReply, error) {
	out := new(PushTaskReply)
	err := c.cc.Invoke(ctx, GoFlow_PushTask_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *goFlowClient) GetResult(ctx context.Context, in *GetResultRequest, opts ...grpc.CallOption) (*GetResultReply, error) {
	out := new(GetResultReply)
	err := c.cc.Invoke(ctx, GoFlow_GetResult_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GoFlowServer is the server API for GoFlow service.
// All implementations must embed UnimplementedGoFlowServer
// for forward compatibility
type GoFlowServer interface {
	PushTask(context.Context, *PushTaskRequest) (*PushTaskReply, error)
	GetResult(context.Context, *GetResultRequest) (*GetResultReply, error)
	mustEmbedUnimplementedGoFlowServer()
}

// UnimplementedGoFlowServer must be embedded to have forward compatible implementations.
type UnimplementedGoFlowServer struct {
}

func (UnimplementedGoFlowServer) PushTask(context.Context, *PushTaskRequest) (*PushTaskReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PushTask not implemented")
}
func (UnimplementedGoFlowServer) GetResult(context.Context, *GetResultRequest) (*GetResultReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResult not implemented")
}
func (UnimplementedGoFlowServer) mustEmbedUnimplementedGoFlowServer() {}

// UnsafeGoFlowServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GoFlowServer will
// result in compilation errors.
type UnsafeGoFlowServer interface {
	mustEmbedUnimplementedGoFlowServer()
}

func RegisterGoFlowServer(s grpc.ServiceRegistrar, srv GoFlowServer) {
	s.RegisterService(&GoFlow_ServiceDesc, srv)
}

func _GoFlow_PushTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoFlowServer).PushTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GoFlow_PushTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoFlowServer).PushTask(ctx, req.(*PushTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GoFlow_GetResult_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetResultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GoFlowServer).GetResult(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GoFlow_GetResult_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GoFlowServer).GetResult(ctx, req.(*GetResultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GoFlow_ServiceDesc is the grpc.ServiceDesc for GoFlow service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GoFlow_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "goflow.GoFlow",
	HandlerType: (*GoFlowServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PushTask",
			Handler:    _GoFlow_PushTask_Handler,
		},
		{
			MethodName: "GetResult",
			Handler:    _GoFlow_GetResult_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/proto/goflow.proto",
}
