// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rest-service.proto

package api_v2

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("rest-service.proto", fileDescriptor_abbfcffb0e8b5cee) }

var fileDescriptor_abbfcffb0e8b5cee = []byte{
	// 200 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x2a, 0x4a, 0x2d, 0x2e,
	0xd1, 0x2d, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62,
	0x2a, 0x33, 0x92, 0x92, 0x49, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x4f, 0x2c, 0xc8, 0xd4, 0x4f,
	0xcc, 0xcb, 0xcb, 0x2f, 0x49, 0x2c, 0xc9, 0xcc, 0xcf, 0x2b, 0x86, 0xa8, 0x90, 0xe2, 0x49, 0xce,
	0xcf, 0xcd, 0xcd, 0xcf, 0x83, 0xf0, 0x8c, 0x66, 0x32, 0x72, 0x71, 0x04, 0x43, 0x4c, 0x70, 0x14,
	0x0a, 0xe4, 0x62, 0x09, 0xc8, 0xcc, 0x4b, 0x17, 0xe2, 0xd7, 0x2b, 0x33, 0xd2, 0xf3, 0x4d, 0x2d,
	0x2e, 0x4e, 0x4c, 0x4f, 0x05, 0x09, 0x48, 0xa1, 0x08, 0xe4, 0xe7, 0xa5, 0x2b, 0xa9, 0x37, 0x5d,
	0x7e, 0x32, 0x99, 0x49, 0x51, 0x48, 0x5e, 0x3f, 0x39, 0xbf, 0x28, 0x55, 0x1f, 0xea, 0x06, 0x47,
	0xfd, 0x82, 0xcc, 0xbc, 0x74, 0xfd, 0xea, 0x92, 0xcc, 0xdc, 0xd4, 0xe2, 0x92, 0xc4, 0xdc, 0x82,
	0x5a, 0x21, 0x2b, 0x2e, 0x96, 0x80, 0xfc, 0xe2, 0x12, 0x22, 0x8c, 0x14, 0x00, 0x1b, 0xc9, 0xa5,
	0xc4, 0xaa, 0x5f, 0x90, 0x5f, 0x5c, 0x62, 0xc5, 0xa8, 0xe5, 0xc4, 0x15, 0xc0, 0x18, 0xc5, 0x96,
	0x58, 0x90, 0xa9, 0x57, 0x66, 0x94, 0xc4, 0x06, 0x76, 0xae, 0x31, 0x20, 0x00, 0x00, 0xff, 0xff,
	0x3e, 0x0f, 0x4a, 0x2c, 0xf4, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ServiceAClient is the client API for ServiceA service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ServiceAClient interface {
	Ping(ctx context.Context, in *MessagePing, opts ...grpc.CallOption) (*MessagePong, error)
	Post(ctx context.Context, in *MessagePing, opts ...grpc.CallOption) (*MessagePong, error)
}

type serviceAClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceAClient(cc grpc.ClientConnInterface) ServiceAClient {
	return &serviceAClient{cc}
}

func (c *serviceAClient) Ping(ctx context.Context, in *MessagePing, opts ...grpc.CallOption) (*MessagePong, error) {
	out := new(MessagePong)
	err := c.cc.Invoke(ctx, "/v2.ServiceA/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceAClient) Post(ctx context.Context, in *MessagePing, opts ...grpc.CallOption) (*MessagePong, error) {
	out := new(MessagePong)
	err := c.cc.Invoke(ctx, "/v2.ServiceA/Post", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ServiceAServer is the server API for ServiceA service.
type ServiceAServer interface {
	Ping(context.Context, *MessagePing) (*MessagePong, error)
	Post(context.Context, *MessagePing) (*MessagePong, error)
}

// UnimplementedServiceAServer can be embedded to have forward compatible implementations.
type UnimplementedServiceAServer struct {
}

func (*UnimplementedServiceAServer) Ping(ctx context.Context, req *MessagePing) (*MessagePong, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedServiceAServer) Post(ctx context.Context, req *MessagePing) (*MessagePong, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Post not implemented")
}

func RegisterServiceAServer(s *grpc.Server, srv ServiceAServer) {
	s.RegisterService(&_ServiceA_serviceDesc, srv)
}

func _ServiceA_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessagePing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceAServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v2.ServiceA/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceAServer).Ping(ctx, req.(*MessagePing))
	}
	return interceptor(ctx, in, info, handler)
}

func _ServiceA_Post_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessagePing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ServiceAServer).Post(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/v2.ServiceA/Post",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ServiceAServer).Post(ctx, req.(*MessagePing))
	}
	return interceptor(ctx, in, info, handler)
}

var _ServiceA_serviceDesc = grpc.ServiceDesc{
	ServiceName: "v2.ServiceA",
	HandlerType: (*ServiceAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _ServiceA_Ping_Handler,
		},
		{
			MethodName: "Post",
			Handler:    _ServiceA_Post_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rest-service.proto",
}
