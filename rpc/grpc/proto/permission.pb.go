// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.7.1
// source: permission.proto

package proto

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
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

type AdminUpdateParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Admin     string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	Successor string `protobuf:"bytes,2,opt,name=successor,proto3" json:"successor,omitempty"`
	Comment   string `protobuf:"bytes,3,opt,name=comment,proto3" json:"comment,omitempty"`
}

func (x *AdminUpdateParam) Reset() {
	*x = AdminUpdateParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_permission_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AdminUpdateParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AdminUpdateParam) ProtoMessage() {}

func (x *AdminUpdateParam) ProtoReflect() protoreflect.Message {
	mi := &file_permission_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AdminUpdateParam.ProtoReflect.Descriptor instead.
func (*AdminUpdateParam) Descriptor() ([]byte, []int) {
	return file_permission_proto_rawDescGZIP(), []int{0}
}

func (x *AdminUpdateParam) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *AdminUpdateParam) GetSuccessor() string {
	if x != nil {
		return x.Successor
	}
	return ""
}

func (x *AdminUpdateParam) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

type AdminUser struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Comment string `protobuf:"bytes,2,opt,name=comment,proto3" json:"comment,omitempty"`
}

func (x *AdminUser) Reset() {
	*x = AdminUser{}
	if protoimpl.UnsafeEnabled {
		mi := &file_permission_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AdminUser) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AdminUser) ProtoMessage() {}

func (x *AdminUser) ProtoReflect() protoreflect.Message {
	mi := &file_permission_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AdminUser.ProtoReflect.Descriptor instead.
func (*AdminUser) Descriptor() ([]byte, []int) {
	return file_permission_proto_rawDescGZIP(), []int{1}
}

func (x *AdminUser) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

func (x *AdminUser) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

type NodeParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Admin   string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	NodeId  string `protobuf:"bytes,2,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	NodeUrl string `protobuf:"bytes,3,opt,name=nodeUrl,proto3" json:"nodeUrl,omitempty"`
	Comment string `protobuf:"bytes,4,opt,name=comment,proto3" json:"comment,omitempty"`
}

func (x *NodeParam) Reset() {
	*x = NodeParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_permission_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeParam) ProtoMessage() {}

func (x *NodeParam) ProtoReflect() protoreflect.Message {
	mi := &file_permission_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeParam.ProtoReflect.Descriptor instead.
func (*NodeParam) Descriptor() ([]byte, []int) {
	return file_permission_proto_rawDescGZIP(), []int{2}
}

func (x *NodeParam) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *NodeParam) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *NodeParam) GetNodeUrl() string {
	if x != nil {
		return x.NodeUrl
	}
	return ""
}

func (x *NodeParam) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

type NodeInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId  string `protobuf:"bytes,1,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	NodeUrl string `protobuf:"bytes,2,opt,name=nodeUrl,proto3" json:"nodeUrl,omitempty"`
	Comment string `protobuf:"bytes,3,opt,name=comment,proto3" json:"comment,omitempty"`
}

func (x *NodeInfo) Reset() {
	*x = NodeInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_permission_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeInfo) ProtoMessage() {}

func (x *NodeInfo) ProtoReflect() protoreflect.Message {
	mi := &file_permission_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeInfo.ProtoReflect.Descriptor instead.
func (*NodeInfo) Descriptor() ([]byte, []int) {
	return file_permission_proto_rawDescGZIP(), []int{3}
}

func (x *NodeInfo) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *NodeInfo) GetNodeUrl() string {
	if x != nil {
		return x.NodeUrl
	}
	return ""
}

func (x *NodeInfo) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

type NodeInfos struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Nodes []*NodeInfo `protobuf:"bytes,1,rep,name=nodes,proto3" json:"nodes,omitempty"`
}

func (x *NodeInfos) Reset() {
	*x = NodeInfos{}
	if protoimpl.UnsafeEnabled {
		mi := &file_permission_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NodeInfos) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NodeInfos) ProtoMessage() {}

func (x *NodeInfos) ProtoReflect() protoreflect.Message {
	mi := &file_permission_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NodeInfos.ProtoReflect.Descriptor instead.
func (*NodeInfos) Descriptor() ([]byte, []int) {
	return file_permission_proto_rawDescGZIP(), []int{4}
}

func (x *NodeInfos) GetNodes() []*NodeInfo {
	if x != nil {
		return x.Nodes
	}
	return nil
}

var File_permission_proto protoreflect.FileDescriptor

var file_permission_proto_rawDesc = []byte{
	0x0a, 0x10, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x62, 0x61, 0x73, 0x69,
	0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x60, 0x0a, 0x10, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x64, 0x6d,
	0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12,
	0x1c, 0x0a, 0x09, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x3f, 0x0a, 0x09, 0x41, 0x64, 0x6d, 0x69, 0x6e,
	0x55, 0x73, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x6d, 0x0a, 0x09, 0x4e, 0x6f, 0x64, 0x65,
	0x50, 0x61, 0x72, 0x61, 0x6d, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6e,
	0x6f, 0x64, 0x65, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64,
	0x65, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x56, 0x0a, 0x08, 0x4e, 0x6f, 0x64, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6e,
	0x6f, 0x64, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x6f,
	0x64, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x22,
	0x32, 0x0a, 0x09, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x12, 0x25, 0x0a, 0x05,
	0x6e, 0x6f, 0x64, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x6e, 0x6f,
	0x64, 0x65, 0x73, 0x32, 0xa3, 0x04, 0x0a, 0x0d, 0x50, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x41, 0x50, 0x49, 0x12, 0x6e, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x48, 0x61, 0x6e, 0x64, 0x6f, 0x76, 0x65, 0x72, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x17,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x23, 0x12, 0x21, 0x2f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f,
	0x67, 0x65, 0x74, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x48, 0x61, 0x6e, 0x64, 0x6f, 0x76, 0x65, 0x72,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x52, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x55, 0x73, 0x65, 0x72, 0x22, 0x1c, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x16, 0x12, 0x14, 0x2f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x2f, 0x67, 0x65, 0x74, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x61, 0x0a, 0x12, 0x47, 0x65, 0x74,
	0x4e, 0x6f, 0x64, 0x65, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12,
	0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x26, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20, 0x12, 0x1e, 0x2f, 0x70,
	0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x67, 0x65, 0x74, 0x4e, 0x6f, 0x64,
	0x65, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x46, 0x0a, 0x07,
	0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x1a, 0x0f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e,
	0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x1b, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x15, 0x12,
	0x13, 0x2f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x2f, 0x67, 0x65, 0x74,
	0x4e, 0x6f, 0x64, 0x65, 0x12, 0x58, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0c, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x22, 0x21, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x1b, 0x12, 0x19, 0x2f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x2f, 0x67, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x49,
	0x0a, 0x08, 0x47, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x12, 0x0d, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x4f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x1a, 0x10, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x4e, 0x6f, 0x64, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x73, 0x22, 0x1c, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x16, 0x12, 0x14, 0x2f, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x2f, 0x67, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x73, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_permission_proto_rawDescOnce sync.Once
	file_permission_proto_rawDescData = file_permission_proto_rawDesc
)

func file_permission_proto_rawDescGZIP() []byte {
	file_permission_proto_rawDescOnce.Do(func() {
		file_permission_proto_rawDescData = protoimpl.X.CompressGZIP(file_permission_proto_rawDescData)
	})
	return file_permission_proto_rawDescData
}

var file_permission_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_permission_proto_goTypes = []interface{}{
	(*AdminUpdateParam)(nil), // 0: proto.AdminUpdateParam
	(*AdminUser)(nil),        // 1: proto.AdminUser
	(*NodeParam)(nil),        // 2: proto.NodeParam
	(*NodeInfo)(nil),         // 3: proto.NodeInfo
	(*NodeInfos)(nil),        // 4: proto.NodeInfos
	(*empty.Empty)(nil),      // 5: google.protobuf.Empty
	(*String)(nil),           // 6: proto.String
	(*Offset)(nil),           // 7: proto.Offset
	(*types.StateBlock)(nil), // 8: types.StateBlock
	(*Int32)(nil),            // 9: proto.Int32
}
var file_permission_proto_depIdxs = []int32{
	3, // 0: proto.NodeInfos.nodes:type_name -> proto.NodeInfo
	0, // 1: proto.PermissionAPI.GetAdminHandoverBlock:input_type -> proto.AdminUpdateParam
	5, // 2: proto.PermissionAPI.GetAdmin:input_type -> google.protobuf.Empty
	2, // 3: proto.PermissionAPI.GetNodeUpdateBlock:input_type -> proto.NodeParam
	6, // 4: proto.PermissionAPI.GetNode:input_type -> proto.String
	5, // 5: proto.PermissionAPI.GetNodesCount:input_type -> google.protobuf.Empty
	7, // 6: proto.PermissionAPI.GetNodes:input_type -> proto.Offset
	8, // 7: proto.PermissionAPI.GetAdminHandoverBlock:output_type -> types.StateBlock
	1, // 8: proto.PermissionAPI.GetAdmin:output_type -> proto.AdminUser
	8, // 9: proto.PermissionAPI.GetNodeUpdateBlock:output_type -> types.StateBlock
	3, // 10: proto.PermissionAPI.GetNode:output_type -> proto.NodeInfo
	9, // 11: proto.PermissionAPI.GetNodesCount:output_type -> proto.Int32
	4, // 12: proto.PermissionAPI.GetNodes:output_type -> proto.NodeInfos
	7, // [7:13] is the sub-list for method output_type
	1, // [1:7] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_permission_proto_init() }
func file_permission_proto_init() {
	if File_permission_proto != nil {
		return
	}
	file_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_permission_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AdminUpdateParam); i {
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
		file_permission_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AdminUser); i {
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
		file_permission_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeParam); i {
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
		file_permission_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeInfo); i {
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
		file_permission_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NodeInfos); i {
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
			RawDescriptor: file_permission_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_permission_proto_goTypes,
		DependencyIndexes: file_permission_proto_depIdxs,
		MessageInfos:      file_permission_proto_msgTypes,
	}.Build()
	File_permission_proto = out.File
	file_permission_proto_rawDesc = nil
	file_permission_proto_goTypes = nil
	file_permission_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// PermissionAPIClient is the client API for PermissionAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PermissionAPIClient interface {
	GetAdminHandoverBlock(ctx context.Context, in *AdminUpdateParam, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetAdmin(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*AdminUser, error)
	GetNodeUpdateBlock(ctx context.Context, in *NodeParam, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetNode(ctx context.Context, in *String, opts ...grpc.CallOption) (*NodeInfo, error)
	GetNodesCount(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Int32, error)
	GetNodes(ctx context.Context, in *Offset, opts ...grpc.CallOption) (*NodeInfos, error)
}

type permissionAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewPermissionAPIClient(cc grpc.ClientConnInterface) PermissionAPIClient {
	return &permissionAPIClient{cc}
}

func (c *permissionAPIClient) GetAdminHandoverBlock(ctx context.Context, in *AdminUpdateParam, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.PermissionAPI/GetAdminHandoverBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *permissionAPIClient) GetAdmin(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*AdminUser, error) {
	out := new(AdminUser)
	err := c.cc.Invoke(ctx, "/proto.PermissionAPI/GetAdmin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *permissionAPIClient) GetNodeUpdateBlock(ctx context.Context, in *NodeParam, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.PermissionAPI/GetNodeUpdateBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *permissionAPIClient) GetNode(ctx context.Context, in *String, opts ...grpc.CallOption) (*NodeInfo, error) {
	out := new(NodeInfo)
	err := c.cc.Invoke(ctx, "/proto.PermissionAPI/GetNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *permissionAPIClient) GetNodesCount(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Int32, error) {
	out := new(Int32)
	err := c.cc.Invoke(ctx, "/proto.PermissionAPI/GetNodesCount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *permissionAPIClient) GetNodes(ctx context.Context, in *Offset, opts ...grpc.CallOption) (*NodeInfos, error) {
	out := new(NodeInfos)
	err := c.cc.Invoke(ctx, "/proto.PermissionAPI/GetNodes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PermissionAPIServer is the server API for PermissionAPI service.
type PermissionAPIServer interface {
	GetAdminHandoverBlock(context.Context, *AdminUpdateParam) (*types.StateBlock, error)
	GetAdmin(context.Context, *empty.Empty) (*AdminUser, error)
	GetNodeUpdateBlock(context.Context, *NodeParam) (*types.StateBlock, error)
	GetNode(context.Context, *String) (*NodeInfo, error)
	GetNodesCount(context.Context, *empty.Empty) (*Int32, error)
	GetNodes(context.Context, *Offset) (*NodeInfos, error)
}

// UnimplementedPermissionAPIServer can be embedded to have forward compatible implementations.
type UnimplementedPermissionAPIServer struct {
}

func (*UnimplementedPermissionAPIServer) GetAdminHandoverBlock(context.Context, *AdminUpdateParam) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAdminHandoverBlock not implemented")
}
func (*UnimplementedPermissionAPIServer) GetAdmin(context.Context, *empty.Empty) (*AdminUser, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAdmin not implemented")
}
func (*UnimplementedPermissionAPIServer) GetNodeUpdateBlock(context.Context, *NodeParam) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNodeUpdateBlock not implemented")
}
func (*UnimplementedPermissionAPIServer) GetNode(context.Context, *String) (*NodeInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNode not implemented")
}
func (*UnimplementedPermissionAPIServer) GetNodesCount(context.Context, *empty.Empty) (*Int32, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNodesCount not implemented")
}
func (*UnimplementedPermissionAPIServer) GetNodes(context.Context, *Offset) (*NodeInfos, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNodes not implemented")
}

func RegisterPermissionAPIServer(s *grpc.Server, srv PermissionAPIServer) {
	s.RegisterService(&_PermissionAPI_serviceDesc, srv)
}

func _PermissionAPI_GetAdminHandoverBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AdminUpdateParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PermissionAPIServer).GetAdminHandoverBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PermissionAPI/GetAdminHandoverBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PermissionAPIServer).GetAdminHandoverBlock(ctx, req.(*AdminUpdateParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _PermissionAPI_GetAdmin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PermissionAPIServer).GetAdmin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PermissionAPI/GetAdmin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PermissionAPIServer).GetAdmin(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _PermissionAPI_GetNodeUpdateBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PermissionAPIServer).GetNodeUpdateBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PermissionAPI/GetNodeUpdateBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PermissionAPIServer).GetNodeUpdateBlock(ctx, req.(*NodeParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _PermissionAPI_GetNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(String)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PermissionAPIServer).GetNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PermissionAPI/GetNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PermissionAPIServer).GetNode(ctx, req.(*String))
	}
	return interceptor(ctx, in, info, handler)
}

func _PermissionAPI_GetNodesCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PermissionAPIServer).GetNodesCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PermissionAPI/GetNodesCount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PermissionAPIServer).GetNodesCount(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _PermissionAPI_GetNodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Offset)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PermissionAPIServer).GetNodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.PermissionAPI/GetNodes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PermissionAPIServer).GetNodes(ctx, req.(*Offset))
	}
	return interceptor(ctx, in, info, handler)
}

var _PermissionAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.PermissionAPI",
	HandlerType: (*PermissionAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAdminHandoverBlock",
			Handler:    _PermissionAPI_GetAdminHandoverBlock_Handler,
		},
		{
			MethodName: "GetAdmin",
			Handler:    _PermissionAPI_GetAdmin_Handler,
		},
		{
			MethodName: "GetNodeUpdateBlock",
			Handler:    _PermissionAPI_GetNodeUpdateBlock_Handler,
		},
		{
			MethodName: "GetNode",
			Handler:    _PermissionAPI_GetNode_Handler,
		},
		{
			MethodName: "GetNodesCount",
			Handler:    _PermissionAPI_GetNodesCount_Handler,
		},
		{
			MethodName: "GetNodes",
			Handler:    _PermissionAPI_GetNodes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "permission.proto",
}
