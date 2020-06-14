// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0
// 	protoc        v3.7.1
// source: representative.proto

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

type RepRewardParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account      string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Beneficial   string `protobuf:"bytes,2,opt,name=beneficial,proto3" json:"beneficial,omitempty"`
	StartHeight  uint64 `protobuf:"varint,3,opt,name=startHeight,proto3" json:"startHeight,omitempty"`
	EndHeight    uint64 `protobuf:"varint,4,opt,name=endHeight,proto3" json:"endHeight,omitempty"`
	RewardBlocks uint64 `protobuf:"varint,5,opt,name=rewardBlocks,proto3" json:"rewardBlocks,omitempty"`
	RewardAmount int64  `protobuf:"varint,6,opt,name=rewardAmount,proto3" json:"rewardAmount,omitempty"`
}

func (x *RepRewardParam) Reset() {
	*x = RepRewardParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_representative_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepRewardParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepRewardParam) ProtoMessage() {}

func (x *RepRewardParam) ProtoReflect() protoreflect.Message {
	mi := &file_representative_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepRewardParam.ProtoReflect.Descriptor instead.
func (*RepRewardParam) Descriptor() ([]byte, []int) {
	return file_representative_proto_rawDescGZIP(), []int{0}
}

func (x *RepRewardParam) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

func (x *RepRewardParam) GetBeneficial() string {
	if x != nil {
		return x.Beneficial
	}
	return ""
}

func (x *RepRewardParam) GetStartHeight() uint64 {
	if x != nil {
		return x.StartHeight
	}
	return 0
}

func (x *RepRewardParam) GetEndHeight() uint64 {
	if x != nil {
		return x.EndHeight
	}
	return 0
}

func (x *RepRewardParam) GetRewardBlocks() uint64 {
	if x != nil {
		return x.RewardBlocks
	}
	return 0
}

func (x *RepRewardParam) GetRewardAmount() int64 {
	if x != nil {
		return x.RewardAmount
	}
	return 0
}

type RepAvailRewardInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LastEndHeight     uint64 `protobuf:"varint,1,opt,name=lastEndHeight,proto3" json:"lastEndHeight,omitempty"`
	LatestBlockHeight uint64 `protobuf:"varint,2,opt,name=latestBlockHeight,proto3" json:"latestBlockHeight,omitempty"`
	NodeRewardHeight  uint64 `protobuf:"varint,3,opt,name=nodeRewardHeight,proto3" json:"nodeRewardHeight,omitempty"`
	AvailStartHeight  uint64 `protobuf:"varint,4,opt,name=availStartHeight,proto3" json:"availStartHeight,omitempty"`
	AvailEndHeight    uint64 `protobuf:"varint,5,opt,name=availEndHeight,proto3" json:"availEndHeight,omitempty"`
	AvailRewardBlocks uint64 `protobuf:"varint,6,opt,name=availRewardBlocks,proto3" json:"availRewardBlocks,omitempty"`
	AvailRewardAmount int64  `protobuf:"varint,7,opt,name=availRewardAmount,proto3" json:"availRewardAmount,omitempty"`
	NeedCallReward    bool   `protobuf:"varint,8,opt,name=needCallReward,proto3" json:"needCallReward,omitempty"`
}

func (x *RepAvailRewardInfo) Reset() {
	*x = RepAvailRewardInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_representative_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepAvailRewardInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepAvailRewardInfo) ProtoMessage() {}

func (x *RepAvailRewardInfo) ProtoReflect() protoreflect.Message {
	mi := &file_representative_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepAvailRewardInfo.ProtoReflect.Descriptor instead.
func (*RepAvailRewardInfo) Descriptor() ([]byte, []int) {
	return file_representative_proto_rawDescGZIP(), []int{1}
}

func (x *RepAvailRewardInfo) GetLastEndHeight() uint64 {
	if x != nil {
		return x.LastEndHeight
	}
	return 0
}

func (x *RepAvailRewardInfo) GetLatestBlockHeight() uint64 {
	if x != nil {
		return x.LatestBlockHeight
	}
	return 0
}

func (x *RepAvailRewardInfo) GetNodeRewardHeight() uint64 {
	if x != nil {
		return x.NodeRewardHeight
	}
	return 0
}

func (x *RepAvailRewardInfo) GetAvailStartHeight() uint64 {
	if x != nil {
		return x.AvailStartHeight
	}
	return 0
}

func (x *RepAvailRewardInfo) GetAvailEndHeight() uint64 {
	if x != nil {
		return x.AvailEndHeight
	}
	return 0
}

func (x *RepAvailRewardInfo) GetAvailRewardBlocks() uint64 {
	if x != nil {
		return x.AvailRewardBlocks
	}
	return 0
}

func (x *RepAvailRewardInfo) GetAvailRewardAmount() int64 {
	if x != nil {
		return x.AvailRewardAmount
	}
	return 0
}

func (x *RepAvailRewardInfo) GetNeedCallReward() bool {
	if x != nil {
		return x.NeedCallReward
	}
	return false
}

type RepHistoryRewardInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LastEndHeight  uint64 `protobuf:"varint,1,opt,name=lastEndHeight,proto3" json:"lastEndHeight,omitempty"`
	RewardBlocks   uint64 `protobuf:"varint,2,opt,name=rewardBlocks,proto3" json:"rewardBlocks,omitempty"`
	RewardAmount   int64  `protobuf:"varint,3,opt,name=rewardAmount,proto3" json:"rewardAmount,omitempty"`
	LastRewardTime int64  `protobuf:"varint,4,opt,name=lastRewardTime,proto3" json:"lastRewardTime,omitempty"`
}

func (x *RepHistoryRewardInfo) Reset() {
	*x = RepHistoryRewardInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_representative_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepHistoryRewardInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepHistoryRewardInfo) ProtoMessage() {}

func (x *RepHistoryRewardInfo) ProtoReflect() protoreflect.Message {
	mi := &file_representative_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepHistoryRewardInfo.ProtoReflect.Descriptor instead.
func (*RepHistoryRewardInfo) Descriptor() ([]byte, []int) {
	return file_representative_proto_rawDescGZIP(), []int{2}
}

func (x *RepHistoryRewardInfo) GetLastEndHeight() uint64 {
	if x != nil {
		return x.LastEndHeight
	}
	return 0
}

func (x *RepHistoryRewardInfo) GetRewardBlocks() uint64 {
	if x != nil {
		return x.RewardBlocks
	}
	return 0
}

func (x *RepHistoryRewardInfo) GetRewardAmount() int64 {
	if x != nil {
		return x.RewardAmount
	}
	return 0
}

func (x *RepHistoryRewardInfo) GetLastRewardTime() int64 {
	if x != nil {
		return x.LastRewardTime
	}
	return 0
}

type RepStateParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	Height  uint64 `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
}

func (x *RepStateParams) Reset() {
	*x = RepStateParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_representative_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepStateParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepStateParams) ProtoMessage() {}

func (x *RepStateParams) ProtoReflect() protoreflect.Message {
	mi := &file_representative_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepStateParams.ProtoReflect.Descriptor instead.
func (*RepStateParams) Descriptor() ([]byte, []int) {
	return file_representative_proto_rawDescGZIP(), []int{3}
}

func (x *RepStateParams) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

func (x *RepStateParams) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

var File_representative_proto protoreflect.FileDescriptor

var file_representative_proto_rawDesc = []byte{
	0x0a, 0x14, 0x72, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x76, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6d,
	0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2f, 0x62, 0x61, 0x73, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2f, 0x70, 0x6f, 0x76, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd2, 0x01,
	0x0a, 0x0e, 0x52, 0x65, 0x70, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d,
	0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x62, 0x65,
	0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x62, 0x65, 0x6e, 0x65, 0x66, 0x69, 0x63, 0x69, 0x61, 0x6c, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x74,
	0x61, 0x72, 0x74, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0b, 0x73, 0x74, 0x61, 0x72, 0x74, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x65, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x09, 0x65, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65,
	0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0c, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x22,
	0x0a, 0x0c, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x41, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x22, 0xec, 0x02, 0x0a, 0x12, 0x52, 0x65, 0x70, 0x41, 0x76, 0x61, 0x69, 0x6c, 0x52,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x24, 0x0a, 0x0d, 0x6c, 0x61, 0x73,
	0x74, 0x45, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x45, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x2c, 0x0a, 0x11, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x11, 0x6c, 0x61, 0x74, 0x65,
	0x73, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x2a, 0x0a,
	0x10, 0x6e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x10, 0x6e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x77,
	0x61, 0x72, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x2a, 0x0a, 0x10, 0x61, 0x76, 0x61,
	0x69, 0x6c, 0x53, 0x74, 0x61, 0x72, 0x74, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x10, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x53, 0x74, 0x61, 0x72, 0x74, 0x48,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x45, 0x6e,
	0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0e, 0x61,
	0x76, 0x61, 0x69, 0x6c, 0x45, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x2c, 0x0a,
	0x11, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x11, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x52,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x2c, 0x0a, 0x11, 0x61,
	0x76, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x77,
	0x61, 0x72, 0x64, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x6e, 0x65, 0x65,
	0x64, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0e, 0x6e, 0x65, 0x65, 0x64, 0x43, 0x61, 0x6c, 0x6c, 0x52, 0x65, 0x77, 0x61, 0x72,
	0x64, 0x22, 0xac, 0x01, 0x0a, 0x14, 0x52, 0x65, 0x70, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79,
	0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x24, 0x0a, 0x0d, 0x6c, 0x61,
	0x73, 0x74, 0x45, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x45, 0x6e, 0x64, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x77, 0x61, 0x72, 0x64, 0x41, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x72, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x41, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x26, 0x0a, 0x0e, 0x6c, 0x61, 0x73, 0x74,
	0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0e, 0x6c, 0x61, 0x73, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x54, 0x69, 0x6d, 0x65,
	0x22, 0x42, 0x0a, 0x0e, 0x52, 0x65, 0x70, 0x53, 0x74, 0x61, 0x74, 0x65, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x68, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x32, 0x88, 0x06, 0x0a, 0x06, 0x52, 0x65, 0x70, 0x41, 0x50, 0x49, 0x12,
	0x50, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x44, 0x61, 0x74, 0x61,
	0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x52, 0x65, 0x77, 0x61,
	0x72, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x42, 0x79, 0x74, 0x65, 0x73, 0x22, 0x1a, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x14, 0x12, 0x12, 0x2f,
	0x72, 0x65, 0x70, 0x2f, 0x67, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x44, 0x61, 0x74,
	0x61, 0x12, 0x56, 0x0a, 0x10, 0x55, 0x6e, 0x70, 0x61, 0x63, 0x6b, 0x52, 0x65, 0x77, 0x61, 0x72,
	0x64, 0x44, 0x61, 0x74, 0x61, 0x12, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x1a, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x52,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x17, 0x12, 0x15, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x75, 0x6e, 0x70, 0x61, 0x63, 0x6b, 0x52,
	0x65, 0x77, 0x61, 0x72, 0x64, 0x44, 0x61, 0x74, 0x61, 0x12, 0x60, 0x0a, 0x12, 0x47, 0x65, 0x74,
	0x41, 0x76, 0x61, 0x69, 0x6c, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x0e, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x1a,
	0x19, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x41, 0x76, 0x61, 0x69, 0x6c,
	0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x1f, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x19, 0x12, 0x17, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x67, 0x65, 0x74, 0x41, 0x76, 0x61, 0x69,
	0x6c, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x5f, 0x0a, 0x12, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x53, 0x65, 0x6e, 0x64, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x52, 0x65, 0x77,
	0x61, 0x72, 0x64, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x1f, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x19, 0x12, 0x17, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x67, 0x65, 0x74, 0x52, 0x65, 0x77,
	0x61, 0x72, 0x64, 0x53, 0x65, 0x6e, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x5e, 0x0a, 0x12,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x63, 0x76, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x12, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x22, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1c,
	0x22, 0x17, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x67, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64,
	0x52, 0x65, 0x63, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x3a, 0x01, 0x2a, 0x12, 0x69, 0x0a, 0x1c,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x63, 0x76, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x42, 0x79, 0x53, 0x65, 0x6e, 0x64, 0x48, 0x61, 0x73, 0x68, 0x12, 0x0b, 0x2e, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x29, 0x82, 0xd3,
	0xe4, 0x93, 0x02, 0x23, 0x12, 0x21, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x67, 0x65, 0x74, 0x52, 0x65,
	0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x63, 0x76, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x42, 0x79, 0x53,
	0x65, 0x6e, 0x64, 0x48, 0x61, 0x73, 0x68, 0x12, 0x66, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x70, 0x53, 0x74, 0x61, 0x74, 0x65, 0x57, 0x69, 0x74, 0x68, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x1a, 0x12, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x50, 0x6f, 0x76, 0x52, 0x65, 0x70, 0x53, 0x74, 0x61, 0x74, 0x65, 0x22, 0x22, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x1c, 0x12, 0x1a, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x67, 0x65, 0x74, 0x52, 0x65, 0x70,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x57, 0x69, 0x74, 0x68, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x5e, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x48, 0x69, 0x73, 0x74,
	0x6f, 0x72, 0x79, 0x12, 0x0e, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x1a, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x52, 0x65, 0x70, 0x48,
	0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x49, 0x6e, 0x66, 0x6f,
	0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x12, 0x15, 0x2f, 0x72, 0x65, 0x70, 0x2f, 0x67,
	0x65, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x42,
	0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_representative_proto_rawDescOnce sync.Once
	file_representative_proto_rawDescData = file_representative_proto_rawDesc
)

func file_representative_proto_rawDescGZIP() []byte {
	file_representative_proto_rawDescOnce.Do(func() {
		file_representative_proto_rawDescData = protoimpl.X.CompressGZIP(file_representative_proto_rawDescData)
	})
	return file_representative_proto_rawDescData
}

var file_representative_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_representative_proto_goTypes = []interface{}{
	(*RepRewardParam)(nil),       // 0: proto.RepRewardParam
	(*RepAvailRewardInfo)(nil),   // 1: proto.RepAvailRewardInfo
	(*RepHistoryRewardInfo)(nil), // 2: proto.RepHistoryRewardInfo
	(*RepStateParams)(nil),       // 3: proto.RepStateParams
	(*Bytes)(nil),                // 4: proto.Bytes
	(*types.Address)(nil),        // 5: types.Address
	(*types.StateBlock)(nil),     // 6: types.StateBlock
	(*types.Hash)(nil),           // 7: types.Hash
	(*types.PovRepState)(nil),    // 8: types.PovRepState
}
var file_representative_proto_depIdxs = []int32{
	0, // 0: proto.RepAPI.GetRewardData:input_type -> proto.RepRewardParam
	4, // 1: proto.RepAPI.UnpackRewardData:input_type -> proto.Bytes
	5, // 2: proto.RepAPI.GetAvailRewardInfo:input_type -> types.Address
	0, // 3: proto.RepAPI.GetRewardSendBlock:input_type -> proto.RepRewardParam
	6, // 4: proto.RepAPI.GetRewardRecvBlock:input_type -> types.StateBlock
	7, // 5: proto.RepAPI.GetRewardRecvBlockBySendHash:input_type -> types.Hash
	3, // 6: proto.RepAPI.GetRepStateWithHeight:input_type -> proto.RepStateParams
	5, // 7: proto.RepAPI.GetRewardHistory:input_type -> types.Address
	4, // 8: proto.RepAPI.GetRewardData:output_type -> proto.Bytes
	0, // 9: proto.RepAPI.UnpackRewardData:output_type -> proto.RepRewardParam
	1, // 10: proto.RepAPI.GetAvailRewardInfo:output_type -> proto.RepAvailRewardInfo
	6, // 11: proto.RepAPI.GetRewardSendBlock:output_type -> types.StateBlock
	6, // 12: proto.RepAPI.GetRewardRecvBlock:output_type -> types.StateBlock
	6, // 13: proto.RepAPI.GetRewardRecvBlockBySendHash:output_type -> types.StateBlock
	8, // 14: proto.RepAPI.GetRepStateWithHeight:output_type -> types.PovRepState
	2, // 15: proto.RepAPI.GetRewardHistory:output_type -> proto.RepHistoryRewardInfo
	8, // [8:16] is the sub-list for method output_type
	0, // [0:8] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_representative_proto_init() }
func file_representative_proto_init() {
	if File_representative_proto != nil {
		return
	}
	file_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_representative_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepRewardParam); i {
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
		file_representative_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepAvailRewardInfo); i {
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
		file_representative_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepHistoryRewardInfo); i {
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
		file_representative_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepStateParams); i {
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
			RawDescriptor: file_representative_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_representative_proto_goTypes,
		DependencyIndexes: file_representative_proto_depIdxs,
		MessageInfos:      file_representative_proto_msgTypes,
	}.Build()
	File_representative_proto = out.File
	file_representative_proto_rawDesc = nil
	file_representative_proto_goTypes = nil
	file_representative_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// RepAPIClient is the client API for RepAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type RepAPIClient interface {
	GetRewardData(ctx context.Context, in *RepRewardParam, opts ...grpc.CallOption) (*Bytes, error)
	UnpackRewardData(ctx context.Context, in *Bytes, opts ...grpc.CallOption) (*RepRewardParam, error)
	GetAvailRewardInfo(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*RepAvailRewardInfo, error)
	GetRewardSendBlock(ctx context.Context, in *RepRewardParam, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetRewardRecvBlock(ctx context.Context, in *types.StateBlock, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetRewardRecvBlockBySendHash(ctx context.Context, in *types.Hash, opts ...grpc.CallOption) (*types.StateBlock, error)
	GetRepStateWithHeight(ctx context.Context, in *RepStateParams, opts ...grpc.CallOption) (*types.PovRepState, error)
	GetRewardHistory(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*RepHistoryRewardInfo, error)
}

type repAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewRepAPIClient(cc grpc.ClientConnInterface) RepAPIClient {
	return &repAPIClient{cc}
}

func (c *repAPIClient) GetRewardData(ctx context.Context, in *RepRewardParam, opts ...grpc.CallOption) (*Bytes, error) {
	out := new(Bytes)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetRewardData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) UnpackRewardData(ctx context.Context, in *Bytes, opts ...grpc.CallOption) (*RepRewardParam, error) {
	out := new(RepRewardParam)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/UnpackRewardData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) GetAvailRewardInfo(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*RepAvailRewardInfo, error) {
	out := new(RepAvailRewardInfo)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetAvailRewardInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) GetRewardSendBlock(ctx context.Context, in *RepRewardParam, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetRewardSendBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) GetRewardRecvBlock(ctx context.Context, in *types.StateBlock, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetRewardRecvBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) GetRewardRecvBlockBySendHash(ctx context.Context, in *types.Hash, opts ...grpc.CallOption) (*types.StateBlock, error) {
	out := new(types.StateBlock)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetRewardRecvBlockBySendHash", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) GetRepStateWithHeight(ctx context.Context, in *RepStateParams, opts ...grpc.CallOption) (*types.PovRepState, error) {
	out := new(types.PovRepState)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetRepStateWithHeight", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repAPIClient) GetRewardHistory(ctx context.Context, in *types.Address, opts ...grpc.CallOption) (*RepHistoryRewardInfo, error) {
	out := new(RepHistoryRewardInfo)
	err := c.cc.Invoke(ctx, "/proto.RepAPI/GetRewardHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RepAPIServer is the server API for RepAPI service.
type RepAPIServer interface {
	GetRewardData(context.Context, *RepRewardParam) (*Bytes, error)
	UnpackRewardData(context.Context, *Bytes) (*RepRewardParam, error)
	GetAvailRewardInfo(context.Context, *types.Address) (*RepAvailRewardInfo, error)
	GetRewardSendBlock(context.Context, *RepRewardParam) (*types.StateBlock, error)
	GetRewardRecvBlock(context.Context, *types.StateBlock) (*types.StateBlock, error)
	GetRewardRecvBlockBySendHash(context.Context, *types.Hash) (*types.StateBlock, error)
	GetRepStateWithHeight(context.Context, *RepStateParams) (*types.PovRepState, error)
	GetRewardHistory(context.Context, *types.Address) (*RepHistoryRewardInfo, error)
}

// UnimplementedRepAPIServer can be embedded to have forward compatible implementations.
type UnimplementedRepAPIServer struct {
}

func (*UnimplementedRepAPIServer) GetRewardData(context.Context, *RepRewardParam) (*Bytes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRewardData not implemented")
}
func (*UnimplementedRepAPIServer) UnpackRewardData(context.Context, *Bytes) (*RepRewardParam, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnpackRewardData not implemented")
}
func (*UnimplementedRepAPIServer) GetAvailRewardInfo(context.Context, *types.Address) (*RepAvailRewardInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailRewardInfo not implemented")
}
func (*UnimplementedRepAPIServer) GetRewardSendBlock(context.Context, *RepRewardParam) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRewardSendBlock not implemented")
}
func (*UnimplementedRepAPIServer) GetRewardRecvBlock(context.Context, *types.StateBlock) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRewardRecvBlock not implemented")
}
func (*UnimplementedRepAPIServer) GetRewardRecvBlockBySendHash(context.Context, *types.Hash) (*types.StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRewardRecvBlockBySendHash not implemented")
}
func (*UnimplementedRepAPIServer) GetRepStateWithHeight(context.Context, *RepStateParams) (*types.PovRepState, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRepStateWithHeight not implemented")
}
func (*UnimplementedRepAPIServer) GetRewardHistory(context.Context, *types.Address) (*RepHistoryRewardInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRewardHistory not implemented")
}

func RegisterRepAPIServer(s *grpc.Server, srv RepAPIServer) {
	s.RegisterService(&_RepAPI_serviceDesc, srv)
}

func _RepAPI_GetRewardData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepRewardParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetRewardData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetRewardData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetRewardData(ctx, req.(*RepRewardParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_UnpackRewardData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Bytes)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).UnpackRewardData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/UnpackRewardData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).UnpackRewardData(ctx, req.(*Bytes))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_GetAvailRewardInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.Address)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetAvailRewardInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetAvailRewardInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetAvailRewardInfo(ctx, req.(*types.Address))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_GetRewardSendBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepRewardParam)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetRewardSendBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetRewardSendBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetRewardSendBlock(ctx, req.(*RepRewardParam))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_GetRewardRecvBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.StateBlock)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetRewardRecvBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetRewardRecvBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetRewardRecvBlock(ctx, req.(*types.StateBlock))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_GetRewardRecvBlockBySendHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.Hash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetRewardRecvBlockBySendHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetRewardRecvBlockBySendHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetRewardRecvBlockBySendHash(ctx, req.(*types.Hash))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_GetRepStateWithHeight_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepStateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetRepStateWithHeight(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetRepStateWithHeight",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetRepStateWithHeight(ctx, req.(*RepStateParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepAPI_GetRewardHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(types.Address)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepAPIServer).GetRewardHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.RepAPI/GetRewardHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepAPIServer).GetRewardHistory(ctx, req.(*types.Address))
	}
	return interceptor(ctx, in, info, handler)
}

var _RepAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.RepAPI",
	HandlerType: (*RepAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRewardData",
			Handler:    _RepAPI_GetRewardData_Handler,
		},
		{
			MethodName: "UnpackRewardData",
			Handler:    _RepAPI_UnpackRewardData_Handler,
		},
		{
			MethodName: "GetAvailRewardInfo",
			Handler:    _RepAPI_GetAvailRewardInfo_Handler,
		},
		{
			MethodName: "GetRewardSendBlock",
			Handler:    _RepAPI_GetRewardSendBlock_Handler,
		},
		{
			MethodName: "GetRewardRecvBlock",
			Handler:    _RepAPI_GetRewardRecvBlock_Handler,
		},
		{
			MethodName: "GetRewardRecvBlockBySendHash",
			Handler:    _RepAPI_GetRewardRecvBlockBySendHash_Handler,
		},
		{
			MethodName: "GetRepStateWithHeight",
			Handler:    _RepAPI_GetRepStateWithHeight_Handler,
		},
		{
			MethodName: "GetRewardHistory",
			Handler:    _RepAPI_GetRewardHistory_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "representative.proto",
}
