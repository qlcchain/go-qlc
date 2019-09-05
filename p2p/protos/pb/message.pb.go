// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

package pb

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type FrontierReq struct {
	Address              []byte   `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`
	Age                  uint32   `protobuf:"varint,2,opt,name=Age,proto3" json:"Age,omitempty"`
	Count                uint32   `protobuf:"varint,3,opt,name=Count,proto3" json:"Count,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FrontierReq) Reset()         { *m = FrontierReq{} }
func (m *FrontierReq) String() string { return proto.CompactTextString(m) }
func (*FrontierReq) ProtoMessage()    {}
func (*FrontierReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{0}
}
func (m *FrontierReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FrontierReq.Unmarshal(m, b)
}
func (m *FrontierReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FrontierReq.Marshal(b, m, deterministic)
}
func (dst *FrontierReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FrontierReq.Merge(dst, src)
}
func (m *FrontierReq) XXX_Size() int {
	return xxx_messageInfo_FrontierReq.Size(m)
}
func (m *FrontierReq) XXX_DiscardUnknown() {
	xxx_messageInfo_FrontierReq.DiscardUnknown(m)
}

var xxx_messageInfo_FrontierReq proto.InternalMessageInfo

func (m *FrontierReq) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *FrontierReq) GetAge() uint32 {
	if m != nil {
		return m.Age
	}
	return 0
}

func (m *FrontierReq) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

type FrontierRsp struct {
	Frontiers            [][]byte `protobuf:"bytes,1,rep,name=Frontiers,proto3" json:"Frontiers,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FrontierRsp) Reset()         { *m = FrontierRsp{} }
func (m *FrontierRsp) String() string { return proto.CompactTextString(m) }
func (*FrontierRsp) ProtoMessage()    {}
func (*FrontierRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{1}
}
func (m *FrontierRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FrontierRsp.Unmarshal(m, b)
}
func (m *FrontierRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FrontierRsp.Marshal(b, m, deterministic)
}
func (dst *FrontierRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FrontierRsp.Merge(dst, src)
}
func (m *FrontierRsp) XXX_Size() int {
	return xxx_messageInfo_FrontierRsp.Size(m)
}
func (m *FrontierRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_FrontierRsp.DiscardUnknown(m)
}

var xxx_messageInfo_FrontierRsp proto.InternalMessageInfo

func (m *FrontierRsp) GetFrontiers() [][]byte {
	if m != nil {
		return m.Frontiers
	}
	return nil
}

type BulkPullReq struct {
	StartHash            []byte   `protobuf:"bytes,1,opt,name=StartHash,proto3" json:"StartHash,omitempty"`
	EndHash              []byte   `protobuf:"bytes,2,opt,name=EndHash,proto3" json:"EndHash,omitempty"`
	PullType             uint32   `protobuf:"varint,3,opt,name=PullType,proto3" json:"PullType,omitempty"`
	Count                uint32   `protobuf:"varint,4,opt,name=Count,proto3" json:"Count,omitempty"`
	Hashes               []byte   `protobuf:"bytes,5,opt,name=Hashes,proto3" json:"Hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BulkPullReq) Reset()         { *m = BulkPullReq{} }
func (m *BulkPullReq) String() string { return proto.CompactTextString(m) }
func (*BulkPullReq) ProtoMessage()    {}
func (*BulkPullReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{2}
}
func (m *BulkPullReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BulkPullReq.Unmarshal(m, b)
}
func (m *BulkPullReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BulkPullReq.Marshal(b, m, deterministic)
}
func (dst *BulkPullReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BulkPullReq.Merge(dst, src)
}
func (m *BulkPullReq) XXX_Size() int {
	return xxx_messageInfo_BulkPullReq.Size(m)
}
func (m *BulkPullReq) XXX_DiscardUnknown() {
	xxx_messageInfo_BulkPullReq.DiscardUnknown(m)
}

var xxx_messageInfo_BulkPullReq proto.InternalMessageInfo

func (m *BulkPullReq) GetStartHash() []byte {
	if m != nil {
		return m.StartHash
	}
	return nil
}

func (m *BulkPullReq) GetEndHash() []byte {
	if m != nil {
		return m.EndHash
	}
	return nil
}

func (m *BulkPullReq) GetPullType() uint32 {
	if m != nil {
		return m.PullType
	}
	return 0
}

func (m *BulkPullReq) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *BulkPullReq) GetHashes() []byte {
	if m != nil {
		return m.Hashes
	}
	return nil
}

type BulkPullRsp struct {
	Blocks               []byte   `protobuf:"bytes,1,opt,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BulkPullRsp) Reset()         { *m = BulkPullRsp{} }
func (m *BulkPullRsp) String() string { return proto.CompactTextString(m) }
func (*BulkPullRsp) ProtoMessage()    {}
func (*BulkPullRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{3}
}
func (m *BulkPullRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BulkPullRsp.Unmarshal(m, b)
}
func (m *BulkPullRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BulkPullRsp.Marshal(b, m, deterministic)
}
func (dst *BulkPullRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BulkPullRsp.Merge(dst, src)
}
func (m *BulkPullRsp) XXX_Size() int {
	return xxx_messageInfo_BulkPullRsp.Size(m)
}
func (m *BulkPullRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_BulkPullRsp.DiscardUnknown(m)
}

var xxx_messageInfo_BulkPullRsp proto.InternalMessageInfo

func (m *BulkPullRsp) GetBlocks() []byte {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type BulkPushBlock struct {
	Blocks               []byte   `protobuf:"bytes,1,opt,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BulkPushBlock) Reset()         { *m = BulkPushBlock{} }
func (m *BulkPushBlock) String() string { return proto.CompactTextString(m) }
func (*BulkPushBlock) ProtoMessage()    {}
func (*BulkPushBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{4}
}
func (m *BulkPushBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BulkPushBlock.Unmarshal(m, b)
}
func (m *BulkPushBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BulkPushBlock.Marshal(b, m, deterministic)
}
func (dst *BulkPushBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BulkPushBlock.Merge(dst, src)
}
func (m *BulkPushBlock) XXX_Size() int {
	return xxx_messageInfo_BulkPushBlock.Size(m)
}
func (m *BulkPushBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_BulkPushBlock.DiscardUnknown(m)
}

var xxx_messageInfo_BulkPushBlock proto.InternalMessageInfo

func (m *BulkPushBlock) GetBlocks() []byte {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type PublishBlock struct {
	Block                []byte   `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PublishBlock) Reset()         { *m = PublishBlock{} }
func (m *PublishBlock) String() string { return proto.CompactTextString(m) }
func (*PublishBlock) ProtoMessage()    {}
func (*PublishBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{5}
}
func (m *PublishBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PublishBlock.Unmarshal(m, b)
}
func (m *PublishBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PublishBlock.Marshal(b, m, deterministic)
}
func (dst *PublishBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PublishBlock.Merge(dst, src)
}
func (m *PublishBlock) XXX_Size() int {
	return xxx_messageInfo_PublishBlock.Size(m)
}
func (m *PublishBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_PublishBlock.DiscardUnknown(m)
}

var xxx_messageInfo_PublishBlock proto.InternalMessageInfo

func (m *PublishBlock) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

type ConfirmReq struct {
	Block                [][]byte `protobuf:"bytes,1,rep,name=block,proto3" json:"block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfirmReq) Reset()         { *m = ConfirmReq{} }
func (m *ConfirmReq) String() string { return proto.CompactTextString(m) }
func (*ConfirmReq) ProtoMessage()    {}
func (*ConfirmReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{6}
}
func (m *ConfirmReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfirmReq.Unmarshal(m, b)
}
func (m *ConfirmReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfirmReq.Marshal(b, m, deterministic)
}
func (dst *ConfirmReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfirmReq.Merge(dst, src)
}
func (m *ConfirmReq) XXX_Size() int {
	return xxx_messageInfo_ConfirmReq.Size(m)
}
func (m *ConfirmReq) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfirmReq.DiscardUnknown(m)
}

var xxx_messageInfo_ConfirmReq proto.InternalMessageInfo

func (m *ConfirmReq) GetBlock() [][]byte {
	if m != nil {
		return m.Block
	}
	return nil
}

type ConfirmAck struct {
	Account              []byte   `protobuf:"bytes,1,opt,name=Account,proto3" json:"Account,omitempty"`
	Signature            []byte   `protobuf:"bytes,2,opt,name=Signature,proto3" json:"Signature,omitempty"`
	Sequence             uint32   `protobuf:"varint,3,opt,name=Sequence,proto3" json:"Sequence,omitempty"`
	Hash                 [][]byte `protobuf:"bytes,4,rep,name=Hash,proto3" json:"Hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfirmAck) Reset()         { *m = ConfirmAck{} }
func (m *ConfirmAck) String() string { return proto.CompactTextString(m) }
func (*ConfirmAck) ProtoMessage()    {}
func (*ConfirmAck) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{7}
}
func (m *ConfirmAck) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfirmAck.Unmarshal(m, b)
}
func (m *ConfirmAck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfirmAck.Marshal(b, m, deterministic)
}
func (dst *ConfirmAck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfirmAck.Merge(dst, src)
}
func (m *ConfirmAck) XXX_Size() int {
	return xxx_messageInfo_ConfirmAck.Size(m)
}
func (m *ConfirmAck) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfirmAck.DiscardUnknown(m)
}

var xxx_messageInfo_ConfirmAck proto.InternalMessageInfo

func (m *ConfirmAck) GetAccount() []byte {
	if m != nil {
		return m.Account
	}
	return nil
}

func (m *ConfirmAck) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *ConfirmAck) GetSequence() uint32 {
	if m != nil {
		return m.Sequence
	}
	return 0
}

func (m *ConfirmAck) GetHash() [][]byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

type PovStatus struct {
	CurrentHeight        uint64   `protobuf:"varint,1,opt,name=CurrentHeight,proto3" json:"CurrentHeight,omitempty"`
	CurrentHash          []byte   `protobuf:"bytes,2,opt,name=CurrentHash,proto3" json:"CurrentHash,omitempty"`
	GenesisHash          []byte   `protobuf:"bytes,3,opt,name=GenesisHash,proto3" json:"GenesisHash,omitempty"`
	CurrentTD            []byte   `protobuf:"bytes,4,opt,name=CurrentTD,proto3" json:"CurrentTD,omitempty"`
	Timestamp            int64    `protobuf:"varint,5,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PovStatus) Reset()         { *m = PovStatus{} }
func (m *PovStatus) String() string { return proto.CompactTextString(m) }
func (*PovStatus) ProtoMessage()    {}
func (*PovStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{8}
}
func (m *PovStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovStatus.Unmarshal(m, b)
}
func (m *PovStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovStatus.Marshal(b, m, deterministic)
}
func (dst *PovStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovStatus.Merge(dst, src)
}
func (m *PovStatus) XXX_Size() int {
	return xxx_messageInfo_PovStatus.Size(m)
}
func (m *PovStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_PovStatus.DiscardUnknown(m)
}

var xxx_messageInfo_PovStatus proto.InternalMessageInfo

func (m *PovStatus) GetCurrentHeight() uint64 {
	if m != nil {
		return m.CurrentHeight
	}
	return 0
}

func (m *PovStatus) GetCurrentHash() []byte {
	if m != nil {
		return m.CurrentHash
	}
	return nil
}

func (m *PovStatus) GetGenesisHash() []byte {
	if m != nil {
		return m.GenesisHash
	}
	return nil
}

func (m *PovStatus) GetCurrentTD() []byte {
	if m != nil {
		return m.CurrentTD
	}
	return nil
}

func (m *PovStatus) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type PovPublishBlock struct {
	Block                []byte   `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PovPublishBlock) Reset()         { *m = PovPublishBlock{} }
func (m *PovPublishBlock) String() string { return proto.CompactTextString(m) }
func (*PovPublishBlock) ProtoMessage()    {}
func (*PovPublishBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{9}
}
func (m *PovPublishBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovPublishBlock.Unmarshal(m, b)
}
func (m *PovPublishBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovPublishBlock.Marshal(b, m, deterministic)
}
func (dst *PovPublishBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovPublishBlock.Merge(dst, src)
}
func (m *PovPublishBlock) XXX_Size() int {
	return xxx_messageInfo_PovPublishBlock.Size(m)
}
func (m *PovPublishBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_PovPublishBlock.DiscardUnknown(m)
}

var xxx_messageInfo_PovPublishBlock proto.InternalMessageInfo

func (m *PovPublishBlock) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

type PovPullBlockReq struct {
	StartHash            []byte   `protobuf:"bytes,1,opt,name=StartHash,proto3" json:"StartHash,omitempty"`
	StartHeight          uint64   `protobuf:"varint,2,opt,name=StartHeight,proto3" json:"StartHeight,omitempty"`
	Count                uint32   `protobuf:"varint,3,opt,name=Count,proto3" json:"Count,omitempty"`
	PullType             uint32   `protobuf:"varint,4,opt,name=PullType,proto3" json:"PullType,omitempty"`
	Reason               uint32   `protobuf:"varint,5,opt,name=Reason,proto3" json:"Reason,omitempty"`
	Locators             []byte   `protobuf:"bytes,6,opt,name=Locators,proto3" json:"Locators,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PovPullBlockReq) Reset()         { *m = PovPullBlockReq{} }
func (m *PovPullBlockReq) String() string { return proto.CompactTextString(m) }
func (*PovPullBlockReq) ProtoMessage()    {}
func (*PovPullBlockReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{10}
}
func (m *PovPullBlockReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovPullBlockReq.Unmarshal(m, b)
}
func (m *PovPullBlockReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovPullBlockReq.Marshal(b, m, deterministic)
}
func (dst *PovPullBlockReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovPullBlockReq.Merge(dst, src)
}
func (m *PovPullBlockReq) XXX_Size() int {
	return xxx_messageInfo_PovPullBlockReq.Size(m)
}
func (m *PovPullBlockReq) XXX_DiscardUnknown() {
	xxx_messageInfo_PovPullBlockReq.DiscardUnknown(m)
}

var xxx_messageInfo_PovPullBlockReq proto.InternalMessageInfo

func (m *PovPullBlockReq) GetStartHash() []byte {
	if m != nil {
		return m.StartHash
	}
	return nil
}

func (m *PovPullBlockReq) GetStartHeight() uint64 {
	if m != nil {
		return m.StartHeight
	}
	return 0
}

func (m *PovPullBlockReq) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *PovPullBlockReq) GetPullType() uint32 {
	if m != nil {
		return m.PullType
	}
	return 0
}

func (m *PovPullBlockReq) GetReason() uint32 {
	if m != nil {
		return m.Reason
	}
	return 0
}

func (m *PovPullBlockReq) GetLocators() []byte {
	if m != nil {
		return m.Locators
	}
	return nil
}

type PovPullBlockRsp struct {
	Count                uint32   `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	Reason               uint32   `protobuf:"varint,3,opt,name=Reason,proto3" json:"Reason,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PovPullBlockRsp) Reset()         { *m = PovPullBlockRsp{} }
func (m *PovPullBlockRsp) String() string { return proto.CompactTextString(m) }
func (*PovPullBlockRsp) ProtoMessage()    {}
func (*PovPullBlockRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_33c57e4bae7b9afd, []int{11}
}
func (m *PovPullBlockRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovPullBlockRsp.Unmarshal(m, b)
}
func (m *PovPullBlockRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovPullBlockRsp.Marshal(b, m, deterministic)
}
func (dst *PovPullBlockRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovPullBlockRsp.Merge(dst, src)
}
func (m *PovPullBlockRsp) XXX_Size() int {
	return xxx_messageInfo_PovPullBlockRsp.Size(m)
}
func (m *PovPullBlockRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_PovPullBlockRsp.DiscardUnknown(m)
}

var xxx_messageInfo_PovPullBlockRsp proto.InternalMessageInfo

func (m *PovPullBlockRsp) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *PovPullBlockRsp) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

func (m *PovPullBlockRsp) GetReason() uint32 {
	if m != nil {
		return m.Reason
	}
	return 0
}

func init() {
	proto.RegisterType((*FrontierReq)(nil), "pb.FrontierReq")
	proto.RegisterType((*FrontierRsp)(nil), "pb.FrontierRsp")
	proto.RegisterType((*BulkPullReq)(nil), "pb.BulkPullReq")
	proto.RegisterType((*BulkPullRsp)(nil), "pb.BulkPullRsp")
	proto.RegisterType((*BulkPushBlock)(nil), "pb.BulkPushBlock")
	proto.RegisterType((*PublishBlock)(nil), "pb.PublishBlock")
	proto.RegisterType((*ConfirmReq)(nil), "pb.ConfirmReq")
	proto.RegisterType((*ConfirmAck)(nil), "pb.ConfirmAck")
	proto.RegisterType((*PovStatus)(nil), "pb.PovStatus")
	proto.RegisterType((*PovPublishBlock)(nil), "pb.PovPublishBlock")
	proto.RegisterType((*PovPullBlockReq)(nil), "pb.PovPullBlockReq")
	proto.RegisterType((*PovPullBlockRsp)(nil), "pb.PovPullBlockRsp")
}

func init() { proto.RegisterFile("message.proto", fileDescriptor_33c57e4bae7b9afd) }

var fileDescriptor_33c57e4bae7b9afd = []byte{
	// 481 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x54, 0xdb, 0x8a, 0xdb, 0x30,
	0x10, 0xc5, 0xb1, 0x93, 0x36, 0x93, 0x98, 0x16, 0xb1, 0x2c, 0xa6, 0xec, 0x43, 0x10, 0x5b, 0x36,
	0x50, 0xe8, 0x4b, 0xbf, 0x20, 0x9b, 0x5e, 0xf6, 0xa1, 0xd0, 0xa0, 0xa4, 0x1f, 0xe0, 0x38, 0x6a,
	0x62, 0xe2, 0x48, 0x5e, 0x49, 0x5e, 0xe8, 0x57, 0xf4, 0x5b, 0xfa, 0xd4, 0xdf, 0x2b, 0x1a, 0xc9,
	0x91, 0x76, 0x69, 0x69, 0xdf, 0x7c, 0x8e, 0x8e, 0x34, 0x67, 0xce, 0x4c, 0x02, 0xf9, 0x89, 0x6b,
	0x5d, 0xee, 0xf9, 0xdb, 0x56, 0x49, 0x23, 0xc9, 0xa0, 0xdd, 0xd2, 0x2f, 0x30, 0xf9, 0xa8, 0xa4,
	0x30, 0x35, 0x57, 0x8c, 0xdf, 0x93, 0x02, 0x9e, 0x2d, 0x76, 0x3b, 0xc5, 0xb5, 0x2e, 0x92, 0x59,
	0x32, 0x9f, 0xb2, 0x1e, 0x92, 0x97, 0x90, 0x2e, 0xf6, 0xbc, 0x18, 0xcc, 0x92, 0x79, 0xce, 0xec,
	0x27, 0xb9, 0x80, 0xe1, 0x52, 0x76, 0xc2, 0x14, 0x29, 0x72, 0x0e, 0xd0, 0x37, 0xd1, 0x83, 0xba,
	0x25, 0x57, 0x30, 0xee, 0xa1, 0x7d, 0x32, 0x9d, 0x4f, 0x59, 0x20, 0xe8, 0x8f, 0x04, 0x26, 0xb7,
	0x5d, 0x73, 0x5c, 0x75, 0x4d, 0x63, 0xcb, 0x5f, 0xc1, 0x78, 0x6d, 0x4a, 0x65, 0xee, 0x4a, 0x7d,
	0xf0, 0x06, 0x02, 0x61, 0xcd, 0x7d, 0x10, 0x3b, 0x3c, 0x1b, 0x38, 0x73, 0x1e, 0x92, 0x57, 0xf0,
	0xdc, 0x3e, 0xb1, 0xf9, 0xde, 0x72, 0xef, 0xe6, 0x8c, 0x83, 0xcd, 0x2c, 0xb2, 0x49, 0x2e, 0x61,
	0x64, 0x6f, 0x72, 0x5d, 0x0c, 0xf1, 0x29, 0x8f, 0xe8, 0xeb, 0xc8, 0x90, 0x6e, 0xad, 0x6c, 0xdb,
	0xc8, 0xea, 0xd8, 0xc7, 0xe1, 0x11, 0xbd, 0x81, 0xdc, 0xc9, 0xf4, 0xe1, 0xd6, 0x32, 0x7f, 0x15,
	0x5e, 0xc3, 0x74, 0xd5, 0x6d, 0x9b, 0xba, 0xd7, 0x5d, 0xc0, 0x10, 0x4f, 0xbc, 0xcc, 0x01, 0x4a,
	0x01, 0x96, 0x52, 0x7c, 0xab, 0xd5, 0xc9, 0xa6, 0x10, 0x69, 0xd2, 0xa0, 0x31, 0x67, 0xcd, 0xa2,
	0x3a, 0xe2, 0xa0, 0xaa, 0x0a, 0xfb, 0xea, 0x07, 0xe5, 0x20, 0x66, 0x58, 0xef, 0x45, 0x69, 0x3a,
	0xc5, 0x7d, 0x4e, 0x81, 0xb0, 0x49, 0xad, 0xf9, 0x7d, 0xc7, 0x45, 0x75, 0x4e, 0xaa, 0xc7, 0x84,
	0x40, 0x86, 0xe1, 0x66, 0x58, 0x16, 0xbf, 0xe9, 0xcf, 0x04, 0xc6, 0x2b, 0xf9, 0xb0, 0x36, 0xa5,
	0xe9, 0x34, 0xb9, 0x86, 0x7c, 0xd9, 0x29, 0xc5, 0x85, 0xb9, 0xe3, 0xf5, 0xfe, 0xe0, 0x6a, 0x67,
	0xec, 0x31, 0x49, 0x66, 0x30, 0xe9, 0x89, 0x30, 0xab, 0x98, 0xb2, 0x8a, 0x4f, 0x5c, 0x70, 0x5d,
	0x6b, 0x54, 0xa4, 0x4e, 0x11, 0x51, 0xb6, 0x0b, 0x7f, 0x61, 0xf3, 0x1e, 0x27, 0x37, 0x65, 0x81,
	0xb0, 0xa7, 0x9b, 0xfa, 0xc4, 0xb5, 0x29, 0x4f, 0x2d, 0x0e, 0x30, 0x65, 0x81, 0xa0, 0x37, 0xf0,
	0x62, 0x25, 0x1f, 0xfe, 0x23, 0xf6, 0x5f, 0x89, 0x57, 0x36, 0x0d, 0xca, 0xfe, 0xbd, 0x82, 0x33,
	0x98, 0x38, 0xe0, 0xda, 0x1f, 0x60, 0xfb, 0x31, 0xf5, 0xe7, 0x5f, 0xc5, 0xa3, 0x05, 0xcd, 0x9e,
	0x2c, 0xe8, 0x25, 0x8c, 0x18, 0x2f, 0xb5, 0x14, 0xd8, 0x49, 0xce, 0x3c, 0xb2, 0x77, 0x3e, 0xcb,
	0xaa, 0x34, 0x52, 0xe9, 0x62, 0x84, 0x46, 0xce, 0x98, 0x7e, 0x7d, 0x62, 0x5c, 0xb7, 0xb6, 0x70,
	0xd8, 0x87, 0x9c, 0x39, 0x10, 0x1a, 0x1f, 0x44, 0x8d, 0x47, 0x25, 0xd3, 0xb8, 0xe4, 0x76, 0x84,
	0x7f, 0x0c, 0xef, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0x9e, 0x53, 0x7a, 0x7a, 0x29, 0x04, 0x00,
	0x00,
}
