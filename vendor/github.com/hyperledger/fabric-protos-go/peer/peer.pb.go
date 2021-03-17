// Code generated by protoc-gen-go. DO NOT EDIT.
// source: peer/peer.proto

package peer

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

func init() { proto.RegisterFile("peer/peer.proto", fileDescriptor_c302117fbb08ad42) }

var fileDescriptor_c302117fbb08ad42 = []byte{
	// 177 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x8f, 0xb1, 0x0a, 0xc2, 0x30,
	0x10, 0x86, 0x37, 0x91, 0x2c, 0x85, 0x0a, 0x22, 0xc5, 0xc9, 0xd9, 0xa6, 0xa0, 0x6f, 0xa0, 0x38,
	0x5b, 0xea, 0xe6, 0x22, 0x6d, 0x73, 0xa6, 0x81, 0x9a, 0x0b, 0x77, 0x75, 0xf0, 0xed, 0xa5, 0xbd,
	0x06, 0x74, 0x49, 0xe0, 0xfb, 0xbf, 0x3b, 0xee, 0x57, 0x49, 0x00, 0xa0, 0x62, 0x7c, 0x74, 0x20,
	0x1c, 0x30, 0x5d, 0x4c, 0x1f, 0x67, 0x2b, 0x09, 0x08, 0x03, 0x72, 0xdd, 0x4b, 0x98, 0x6d, 0xff,
	0xe0, 0x83, 0x80, 0x03, 0x7a, 0x06, 0x49, 0x0f, 0x57, 0xb5, 0xbc, 0x78, 0x83, 0xc4, 0x40, 0xe9,
	0x59, 0x25, 0x25, 0x61, 0x0b, 0xcc, 0xe5, 0x6c, 0xa7, 0x6b, 0xd1, 0x58, 0xdf, 0x9c, 0xf5, 0x60,
	0x22, 0xcf, 0x36, 0x91, 0x47, 0x52, 0xcd, 0x6b, 0x4f, 0x95, 0xda, 0x21, 0x59, 0xdd, 0x7d, 0x02,
	0x50, 0x0f, 0xc6, 0x02, 0xe9, 0x67, 0xdd, 0x90, 0x6b, 0xe3, 0xc4, 0x78, 0xce, 0x7d, 0x6f, 0xdd,
	0xd0, 0xbd, 0x1b, 0xdd, 0xe2, 0xab, 0xf8, 0x51, 0x0b, 0x51, 0x73, 0x51, 0x73, 0x8b, 0x53, 0xcb,
	0x46, 0xfa, 0x1d, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xd5, 0x12, 0xef, 0x66, 0xf9, 0x00, 0x00,
	0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// EndorserClient is the client API for Endorser service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type EndorserClient interface {
	ProcessProposal(ctx context.Context, in *SignedProposal, opts ...grpc.CallOption) (*ProposalResponse, error)
}

type endorserClient struct {
	cc *grpc.ClientConn
}

func NewEndorserClient(cc *grpc.ClientConn) EndorserClient {
	return &endorserClient{cc}
}

func (c *endorserClient) ProcessProposal(ctx context.Context, in *SignedProposal, opts ...grpc.CallOption) (*ProposalResponse, error) {
	out := new(ProposalResponse)
	err := c.cc.Invoke(ctx, "/protos.Endorser/ProcessProposal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EndorserServer is the server API for Endorser service.
type EndorserServer interface {
	ProcessProposal(context.Context, *SignedProposal) (*ProposalResponse, error)
}

// UnimplementedEndorserServer can be embedded to have forward compatible implementations.
type UnimplementedEndorserServer struct {
}

func (*UnimplementedEndorserServer) ProcessProposal(ctx context.Context, req *SignedProposal) (*ProposalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessProposal not implemented")
}

func RegisterEndorserServer(s *grpc.Server, srv EndorserServer) {
	s.RegisterService(&_Endorser_serviceDesc, srv)
}

func _Endorser_ProcessProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignedProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EndorserServer).ProcessProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Endorser/ProcessProposal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EndorserServer).ProcessProposal(ctx, req.(*SignedProposal))
	}
	return interceptor(ctx, in, info, handler)
}

var _Endorser_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Endorser",
	HandlerType: (*EndorserServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessProposal",
			Handler:    _Endorser_ProcessProposal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "peer/peer.proto",
}
