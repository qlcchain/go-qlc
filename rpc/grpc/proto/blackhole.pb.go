// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0
// 	protoc        v3.7.1
// source: blackhole.proto

package proto

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	types "github.com/qlcchain/go-qlc/rpc/grpc/proto/types"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type DestroyParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Owner    string `protobuf:"bytes,1,opt,name=Owner,proto3" json:"Owner,omitempty"`
	Previous string `protobuf:"bytes,2,opt,name=Previous,proto3" json:"Previous,omitempty"`
	Token    string `protobuf:"bytes,3,opt,name=Token,proto3" json:"Token,omitempty"`
	Amount   int64  `protobuf:"varint,4,opt,name=Amount,proto3" json:"Amount,omitempty"`
	Sign     string `protobuf:"bytes,5,opt,name=Sign,proto3" json:"Sign,omitempty"`
}

func (x *DestroyParam) Reset() {
	*x = DestroyParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blackhole_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DestroyParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DestroyParam) ProtoMessage() {}

func (x *DestroyParam) ProtoReflect() protoreflect.Message {
	mi := &file_blackhole_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DestroyParam.ProtoReflect.Descriptor instead.
func (*DestroyParam) Descriptor() ([]byte, []int) {
	return file_blackhole_proto_rawDescGZIP(), []int{0}
}

func (x *DestroyParam) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *DestroyParam) GetPrevious() string {
	if x != nil {
		return x.Previous
	}
	return ""
}

func (x *DestroyParam) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *DestroyParam) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *DestroyParam) GetSign() string {
	if x != nil {
		return x.Sign
	}
	return ""
}

var File_blackhole_proto protoreflect.FileDescriptor

var file_blackhole_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x62, 0x6c, 0x61, 0x63, 0x6b, 0x68, 0x6f, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x62, 0x61,
	0x73, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x14, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x82, 0x01, 0x0a, 0x0c, 0x44, 0x65, 0x73, 0x74, 0x72, 0x6f, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x12, 0x14, 0x0a, 0x05, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f,
	0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x50, 0x72, 0x65, 0x76, 0x69, 0x6f,
	0x75, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x41, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x53, 0x69, 0x67, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x53, 0x69, 0x67, 0x6e, 0x32, 0x8f, 0x03, 0x0a, 0x0c, 0x42, 0x6c, 0x61, 0x63, 0x6b, 0x48, 0x6f,
	0x6c, 0x65, 0x41, 0x50, 0x49, 0x12, 0x5a, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x53, 0x65, 0x6e, 0x64,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x13, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x65,
	0x73, 0x74, 0x72, 0x6f, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x22, 0x82,
	0xd3, 0xe4, 0x93, 0x02, 0x1c, 0x22, 0x17, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x68, 0x6f, 0x6c,
	0x65, 0x2f, 0x67, 0x65, 0x74, 0x73, 0x65, 0x6e, 0x64, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x3a, 0x01,
	0x2a, 0x12, 0x58, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x73, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x0b, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x48, 0x61, 0x73,
	0x68, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x22, 0x1a, 0x2f, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x68, 0x6f, 0x6c, 0x65, 0x2f, 0x67, 0x65, 0x74, 0x72, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x73, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x3a, 0x01, 0x2a, 0x12, 0x60, 0x0a, 0x13, 0x47,
	0x65, 0x74, 0x54, 0x6f, 0x74, 0x61, 0x6c, 0x44, 0x65, 0x73, 0x74, 0x72, 0x6f, 0x79, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x0e, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x1a, 0x0e, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x42, 0x61, 0x6c, 0x61, 0x6e,
	0x63, 0x65, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23, 0x22, 0x1e, 0x2f, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x68, 0x6f, 0x6c, 0x65, 0x2f, 0x67, 0x65, 0x74, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x64,
	0x65, 0x73, 0x74, 0x72, 0x6f, 0x79, 0x69, 0x6e, 0x66, 0x6f, 0x3a, 0x01, 0x2a, 0x12, 0x67, 0x0a,
	0x14, 0x47, 0x65, 0x74, 0x44, 0x65, 0x73, 0x74, 0x72, 0x6f, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x44,
	0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x0e, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x1a, 0x13, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x44, 0x65,
	0x73, 0x74, 0x72, 0x6f, 0x79, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x22, 0x2a, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x24, 0x22, 0x1f, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x68, 0x6f, 0x6c, 0x65, 0x2f, 0x67,
	0x65, 0x74, 0x64, 0x65, 0x73, 0x74, 0x72, 0x6f, 0x79, 0x69, 0x6e, 0x66, 0x6f, 0x64, 0x65, 0x74,
	0x61, 0x69, 0x6c, 0x3a, 0x01, 0x2a, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_blackhole_proto_rawDescOnce sync.Once
	file_blackhole_proto_rawDescData = file_blackhole_proto_rawDesc
)

func file_blackhole_proto_rawDescGZIP() []byte {
	file_blackhole_proto_rawDescOnce.Do(func() {
		file_blackhole_proto_rawDescData = protoimpl.X.CompressGZIP(file_blackhole_proto_rawDescData)
	})
	return file_blackhole_proto_rawDescData
}

var file_blackhole_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_blackhole_proto_goTypes = []interface{}{
	(*DestroyParam)(nil),       // 0: proto.DestroyParam
	(*types.Hash)(nil),         // 1: types.Hash
	(*types.Address)(nil),      // 2: types.Address
	(*types.StateBlock)(nil),   // 3: types.StateBlock
	(*types.Balance)(nil),      // 4: types.Balance
	(*types.DestroyInfos)(nil), // 5: types.DestroyInfos
}
var file_blackhole_proto_depIdxs = []int32{
	0, // 0: proto.BlackHoleAPI.GetSendBlock:input_type -> proto.DestroyParam
	1, // 1: proto.BlackHoleAPI.GetRewardsBlock:input_type -> types.Hash
	2, // 2: proto.BlackHoleAPI.GetTotalDestroyInfo:input_type -> types.Address
	2, // 3: proto.BlackHoleAPI.GetDestroyInfoDetail:input_type -> types.Address
	3, // 4: proto.BlackHoleAPI.GetSendBlock:output_type -> types.StateBlock
	3, // 5: proto.BlackHoleAPI.GetRewardsBlock:output_type -> types.StateBlock
	4, // 6: proto.BlackHoleAPI.GetTotalDestroyInfo:output_type -> types.Balance
	5, // 7: proto.BlackHoleAPI.GetDestroyInfoDetail:output_type -> types.DestroyInfos
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_blackhole_proto_init() }
func file_blackhole_proto_init() {
	if File_blackhole_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_blackhole_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DestroyParam); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_blackhole_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_blackhole_proto_goTypes,
		DependencyIndexes: file_blackhole_proto_depIdxs,
		MessageInfos:      file_blackhole_proto_msgTypes,
	}.Build()
	File_blackhole_proto = out.File
	file_blackhole_proto_rawDesc = nil
	file_blackhole_proto_goTypes = nil
	file_blackhole_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// BlackHoleAPIClient is the client API for BlackHoleAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BlackHoleAPIClient interface {
	GetSendBlock(ctx context.Context, in *DestroyParam, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetRewardsBlock(ctx context.Context, in *types.Hash, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetTotalDestroyInfo(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*types.Balance, error)
	GetDestroyInfoDetail(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*types.DestroyInfos, error)
}

type blackHoleAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewBlackHoleAPIClient(cc grpc.ClientConnInterface) BlackHoleAPIClient {
	return &blackHoleAPIClient{cc}
}

func (c *blackHoleAPIClient) GetSendBlock(ctx context.Context, in *DestroyParam, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.BlackHoleAPI/GetSendBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blackHoleAPIClient) GetRewardsBlock(ctx context.Context, in *types.Hash, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.BlackHoleAPI/GetRewardsBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blackHoleAPIClient) GetTotalDestroyInfo(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*types.Balance, error) {
	out := new(types.Balance)
	err := c.cc.Invoke(ctx, "/proto.BlackHoleAPI/GetTotalDestroyInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blackHoleAPIClient) GetDestroyInfoDetail(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*types.DestroyInfos, error) {
	out := new(types.DestroyInfos)
	err := c.cc.Invoke(ctx, "/proto.BlackHoleAPI/GetDestroyInfoDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlackHoleAPIServer is the server API for BlackHoleAPI service.
type BlackHoleAPIServer interface {
	GetSendBlock(context.Context, *DestroyParam) (*types.StateBlock, error)
	GetRewardsBlock(context.Context, *types.Hash) (*types.StateBlock, error)
	GetTotalDestroyInfo(context.Context, *types.Address) (*types.Balance, error)
	GetDestroyInfoDetail(context.Context, *types.Address) (*types.DestroyInfos, error)
}

// UnimplementedBlackHoleAPIServer can be embedded to have forward compatible implementations.
type UnimplementedBlackHoleAPIServer struct {
}

func (*UnimplementedBlackHoleAPIServer) GetSendBlock(context.Context, *DestroyParam) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSendBlock not implemented")
}
func (*UnimplementedBlackHoleAPIServer) GetRewardsBlock(context.Context, *types.Hash) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRewardsBlock not implemented")
}
func (*UnimplementedBlackHoleAPIServer) GetTotalDestroyInfo(context.Context, *types.Address) (*types.Balance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTotalDestroyInfo not implemented")
}
func (*UnimplementedBlackHoleAPIServer) GetDestroyInfoDetail(context.Context, *types.Address) (*types.DestroyInfos, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDestroyInfoDetail not implemented")
}

func RegisterBlackHoleAPIServer(s *grpc.Server, srv BlackHoleAPIServer) {
	s.RegisterService(&_BlackHoleAPI_serviceDesc, srv)
}

func _BlackHoleAPI_GetSendBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DestroyParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlackHoleAPIServer).GetSendBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.BlackHoleAPI/GetSendBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlackHoleAPIServer).GetSendBlock(ctx, req.(*DestroyParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlackHoleAPI_GetRewardsBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.Hash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlackHoleAPIServer).GetRewardsBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.BlackHoleAPI/GetRewardsBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlackHoleAPIServer).GetRewardsBlock(ctx, req.(*types.Hash))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlackHoleAPI_GetTotalDestroyInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.Address)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlackHoleAPIServer).GetTotalDestroyInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.BlackHoleAPI/GetTotalDestroyInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlackHoleAPIServer).GetTotalDestroyInfo(ctx, req.(*types.Address))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlackHoleAPI_GetDestroyInfoDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.Address)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlackHoleAPIServer).GetDestroyInfoDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.BlackHoleAPI/GetDestroyInfoDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlackHoleAPIServer).GetDestroyInfoDetail(ctx, req.(*types.Address))
	}
	return interceptor(ctx, in, info, handler)
}

var _BlackHoleAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.BlackHoleAPI",
	HandlerType: (*BlackHoleAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSendBlock",
			Handler:    _BlackHoleAPI_GetSendBlock_Handler,
		},
		{
			MethodName: "GetRewardsBlock",
			Handler:    _BlackHoleAPI_GetRewardsBlock_Handler,
		},
		{
			MethodName: "GetTotalDestroyInfo",
			Handler:    _BlackHoleAPI_GetTotalDestroyInfo_Handler,
		},
		{
			MethodName: "GetDestroyInfoDetail",
			Handler:    _BlackHoleAPI_GetDestroyInfoDetail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "blackhole.proto",
}
