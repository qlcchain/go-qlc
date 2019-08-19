// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

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
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
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

func (m *PublishBlock) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

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
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
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

func (m *PovPublishBlock) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

func (m *PovPublishBlock) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

type PovPullBlockReq struct {
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	StartHash            []byte   `protobuf:"bytes,2,opt,name=StartHash,proto3" json:"StartHash,omitempty"`
	StartHeight          uint64   `protobuf:"varint,3,opt,name=StartHeight,proto3" json:"StartHeight,omitempty"`
	Count                uint32   `protobuf:"varint,4,opt,name=Count,proto3" json:"Count,omitempty"`
	PullType             uint32   `protobuf:"varint,5,opt,name=PullType,proto3" json:"PullType,omitempty"`
	Reason               uint32   `protobuf:"varint,6,opt,name=Reason,proto3" json:"Reason,omitempty"`
	Locators             []byte   `protobuf:"bytes,7,opt,name=Locators,proto3" json:"Locators,omitempty"`
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

func (m *PovPullBlockReq) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

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
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	Count                uint32   `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	Block                []byte   `protobuf:"bytes,3,opt,name=block,proto3" json:"block,omitempty"`
	Reason               uint32   `protobuf:"varint,4,opt,name=Reason,proto3" json:"Reason,omitempty"`
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

func (m *PovPullBlockRsp) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

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

<<<<<<< HEAD
func init() { proto.RegisterFile("p2p/protos/pb/message.proto", fileDescriptor_eb45e02bf5f44afc) }

var fileDescriptor_eb45e02bf5f44afc = []byte{
	// 544 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x96, 0xeb, 0x24, 0x25, 0x93, 0x44, 0x54, 0x2b, 0x54, 0x59, 0xd0, 0x43, 0x64, 0x71, 0xa8,
	0x38, 0x10, 0x09, 0x9e, 0x20, 0x09, 0x85, 0x1c, 0x10, 0x8d, 0x9c, 0x5c, 0x38, 0xda, 0xce, 0x90,
	0x58, 0xb1, 0xbd, 0xee, 0xee, 0xba, 0x52, 0x9f, 0x82, 0x67, 0xe1, 0x61, 0x78, 0x1f, 0xb4, 0xb3,
	0x6b, 0x7b, 0x1b, 0x04, 0x02, 0xf5, 0xe6, 0xef, 0xdb, 0xdd, 0x6f, 0xe6, 0x9b, 0x1f, 0xc3, 0xab,
	0xea, 0x5d, 0x35, 0xab, 0x04, 0x57, 0x5c, 0xce, 0xaa, 0x64, 0x56, 0xa0, 0x94, 0xf1, 0x1e, 0xdf,
	0x12, 0xc3, 0xce, 0xaa, 0x24, 0xbc, 0x85, 0xd1, 0x47, 0xc1, 0x4b, 0x95, 0xa1, 0x88, 0xf0, 0x8e,
	0x05, 0x70, 0x3e, 0xdf, 0xed, 0x04, 0x4a, 0x19, 0x78, 0x53, 0xef, 0x7a, 0x1c, 0x35, 0x90, 0x5d,
	0x80, 0x3f, 0xdf, 0x63, 0x70, 0x36, 0xf5, 0xae, 0x27, 0x91, 0xfe, 0x64, 0x2f, 0xa0, 0xbf, 0xe4,
	0x75, 0xa9, 0x02, 0x9f, 0x38, 0x03, 0xc2, 0x07, 0x47, 0x50, 0x56, 0xec, 0x0d, 0x5c, 0x6c, 0xb9,
	0x8a, 0xf3, 0x86, 0xfb, 0x52, 0x17, 0xa4, 0x3c, 0x89, 0x7e, 0xe3, 0xd9, 0x14, 0x46, 0x2b, 0x8c,
	0x77, 0x28, 0x16, 0x39, 0x4f, 0x8f, 0x14, 0x6a, 0x1c, 0xb9, 0x14, 0xbb, 0x82, 0xe1, 0x6d, 0x85,
	0xa5, 0x39, 0xf7, 0xe9, 0xbc, 0x23, 0xc2, 0xef, 0x1e, 0x8c, 0x16, 0x75, 0x7e, 0x5c, 0xd7, 0x79,
	0xae, 0xcd, 0x5c, 0xc1, 0x70, 0xa3, 0x62, 0xa1, 0x56, 0xb1, 0x3c, 0x58, 0x3b, 0x1d, 0xa1, 0xad,
	0xde, 0x94, 0x3b, 0x3a, 0x33, 0x91, 0x1a, 0xc8, 0x5e, 0xc2, 0x33, 0x2d, 0xb1, 0x7d, 0xa8, 0xd0,
	0x7a, 0x6b, 0x71, 0x67, 0xba, 0xe7, 0x98, 0x66, 0x97, 0x30, 0xd0, 0x2f, 0x51, 0x06, 0x7d, 0x92,
	0xb2, 0x28, 0xfc, 0xea, 0x24, 0x24, 0x2b, 0x9d, 0x50, 0xa2, 0x33, 0x55, 0x5a, 0xd9, 0x54, 0xa1,
	0x23, 0xb4, 0x74, 0xe2, 0x18, 0x37, 0x40, 0x4b, 0xd3, 0x87, 0xb4, 0x7e, 0x2d, 0x0a, 0x97, 0x30,
	0x31, 0xd2, 0xf2, 0xd0, 0xd6, 0xe6, 0x7f, 0xc5, 0xc3, 0x05, 0x8c, 0xd7, 0x75, 0x92, 0x67, 0x4f,
	0xd1, 0x08, 0x01, 0x96, 0xbc, 0xfc, 0x96, 0x89, 0x42, 0xd7, 0xbc, 0xbd, 0xe3, 0x4d, 0xfd, 0xee,
	0x8e, 0x6a, 0xef, 0xcc, 0xd3, 0x23, 0x0d, 0x59, 0x9a, 0x52, 0x15, 0x9b, 0x21, 0x33, 0x90, 0x3a,
	0x96, 0xed, 0xcb, 0x58, 0xd5, 0x02, 0x6d, 0x94, 0x8e, 0xd0, 0x7d, 0xd9, 0xe0, 0x5d, 0x8d, 0x65,
	0xda, 0xf6, 0xa5, 0xc1, 0x8c, 0x41, 0x8f, 0x5a, 0xd9, 0xa3, 0xb0, 0xf4, 0x1d, 0xfe, 0xf0, 0x60,
	0xb8, 0xe6, 0xf7, 0x1b, 0x15, 0xab, 0x5a, 0xb2, 0xd7, 0x30, 0x59, 0xd6, 0x42, 0x60, 0xa9, 0x56,
	0x98, 0xed, 0x0f, 0x26, 0x76, 0x2f, 0x7a, 0x4c, 0xea, 0x19, 0x6c, 0x88, 0x6e, 0x32, 0x5c, 0x4a,
	0xdf, 0xf8, 0x84, 0x25, 0xca, 0x4c, 0xd2, 0x0d, 0xd3, 0x15, 0x97, 0xd2, 0x2e, 0xec, 0x83, 0xed,
	0x07, 0x9a, 0x93, 0x71, 0xd4, 0x11, 0xfa, 0x74, 0x9b, 0x15, 0x28, 0x55, 0x5c, 0x54, 0x34, 0x2e,
	0x7e, 0xd4, 0x11, 0xe1, 0x0d, 0x3c, 0x5f, 0xf3, 0xfb, 0x27, 0x37, 0xe5, 0xa7, 0x67, 0x75, 0xf2,
	0x9c, 0x44, 0xec, 0x3a, 0xfc, 0x45, 0xe7, 0xd1, 0xb2, 0x9c, 0x9d, 0x2e, 0xcb, 0x14, 0x46, 0x06,
	0x98, 0xd2, 0xf9, 0x54, 0x3a, 0x97, 0xfa, 0xc3, 0x62, 0xb8, 0xab, 0xd4, 0x3f, 0x59, 0xa5, 0x4b,
	0x18, 0x44, 0x18, 0x4b, 0x5e, 0x06, 0x03, 0x3a, 0xb1, 0x48, 0xbf, 0xf9, 0xcc, 0xd3, 0x58, 0x71,
	0x21, 0x83, 0x73, 0x4a, 0xa4, 0xc5, 0xa1, 0x3c, 0xb1, 0xf5, 0x2f, 0x4b, 0x65, 0x26, 0xcd, 0xfc,
	0xb8, 0x0c, 0xe8, 0x8a, 0xe6, 0x9f, 0xac, 0x9a, 0x4d, 0xa8, 0xe7, 0x26, 0x94, 0x0c, 0xe8, 0x77,
	0xf9, 0xfe, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0xa1, 0xc4, 0x8b, 0x16, 0x4d, 0x05, 0x00, 0x00,
=======
func init() { proto.RegisterFile("message.proto", fileDescriptor_33c57e4bae7b9afd) }

var fileDescriptor_33c57e4bae7b9afd = []byte{
	// 498 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0x51, 0x8e, 0xd3, 0x30,
	0x14, 0x54, 0x9a, 0xb4, 0x4b, 0x5f, 0x5b, 0x81, 0x2c, 0xb4, 0x8a, 0x50, 0x3f, 0x2a, 0x0b, 0x44,
	0x25, 0x24, 0x7e, 0xb8, 0x00, 0x6d, 0x59, 0xd8, 0x0f, 0x24, 0xaa, 0xb4, 0x17, 0x48, 0x53, 0xd3,
	0x46, 0x4d, 0xed, 0xac, 0xed, 0xac, 0xc4, 0x29, 0x38, 0x0b, 0x87, 0xe1, 0x3e, 0xc8, 0xcf, 0x4e,
	0xec, 0xad, 0xb4, 0x20, 0xb4, 0x7f, 0x99, 0xf1, 0xcb, 0x64, 0x3c, 0x63, 0x07, 0x26, 0x67, 0xa6,
	0x54, 0x7e, 0x60, 0xef, 0x6b, 0x29, 0xb4, 0x20, 0xbd, 0x7a, 0x47, 0xbf, 0xc1, 0xe8, 0xb3, 0x14,
	0x5c, 0x97, 0x4c, 0x66, 0xec, 0x8e, 0xa4, 0x70, 0xb5, 0xd8, 0xef, 0x25, 0x53, 0x2a, 0x8d, 0x66,
	0xd1, 0x7c, 0x9c, 0xb5, 0x90, 0xbc, 0x80, 0x78, 0x71, 0x60, 0x69, 0x6f, 0x16, 0xcd, 0x27, 0x99,
	0x79, 0x24, 0x2f, 0xa1, 0xbf, 0x12, 0x0d, 0xd7, 0x69, 0x8c, 0x9c, 0x05, 0xf4, 0x5d, 0x20, 0xa8,
	0x6a, 0x32, 0x85, 0x61, 0x0b, 0x8d, 0x64, 0x3c, 0x1f, 0x67, 0x9e, 0xa0, 0x3f, 0x23, 0x18, 0x2d,
	0x9b, 0xea, 0xb4, 0x6e, 0xaa, 0xca, 0x7c, 0x7e, 0x0a, 0xc3, 0x8d, 0xce, 0xa5, 0xbe, 0xcd, 0xd5,
	0xd1, 0x19, 0xf0, 0x84, 0x31, 0x77, 0xc3, 0xf7, 0xb8, 0xd6, 0xb3, 0xe6, 0x1c, 0x24, 0xaf, 0xe0,
	0x99, 0x91, 0xd8, 0xfe, 0xa8, 0x99, 0x73, 0xd3, 0x61, 0x6f, 0x33, 0x09, 0x6c, 0x92, 0x6b, 0x18,
	0x98, 0x37, 0x99, 0x4a, 0xfb, 0x28, 0xe5, 0x10, 0x7d, 0x13, 0x18, 0x52, 0xb5, 0x19, 0xdb, 0x55,
	0xa2, 0x38, 0xb5, 0x71, 0x38, 0x44, 0xdf, 0xc2, 0xc4, 0x8e, 0xa9, 0xe3, 0xd2, 0x30, 0x8f, 0x0e,
	0x2e, 0x61, 0xbc, 0x6e, 0x76, 0x55, 0xd9, 0xce, 0x4d, 0x61, 0x88, 0x2b, 0xda, 0x58, 0x8d, 0xd0,
	0x91, 0x27, 0x8c, 0x57, 0x04, 0x6e, 0x7f, 0x16, 0xd0, 0x8f, 0x00, 0x2b, 0xc1, 0xbf, 0x97, 0xf2,
	0xec, 0x32, 0xfa, 0x6f, 0x05, 0xdd, 0x29, 0x2c, 0x8a, 0x13, 0x96, 0x5c, 0x14, 0x98, 0x49, 0x5b,
	0xb2, 0x85, 0x98, 0x7f, 0x79, 0xe0, 0xb9, 0x6e, 0x24, 0x73, 0x0a, 0x9e, 0x30, 0x29, 0x6f, 0xd8,
	0x5d, 0xc3, 0x78, 0xd1, 0xa5, 0xdc, 0x62, 0x42, 0x20, 0xc1, 0x62, 0x12, 0xac, 0x18, 0x9f, 0xe9,
	0xaf, 0x08, 0x86, 0x6b, 0x71, 0xbf, 0xd1, 0xb9, 0x6e, 0x14, 0x79, 0x0d, 0x93, 0x55, 0x23, 0x25,
	0xe3, 0xfa, 0x96, 0x95, 0x87, 0xa3, 0xfd, 0x76, 0x92, 0x3d, 0x24, 0xc9, 0x0c, 0x46, 0x2d, 0xe1,
	0x7b, 0x0e, 0x29, 0x33, 0xf1, 0x85, 0x71, 0xa6, 0x4a, 0x85, 0x13, 0xb1, 0x9d, 0x08, 0x28, 0xb3,
	0x0b, 0xf7, 0xc2, 0xf6, 0x13, 0xb6, 0x3e, 0xce, 0x3c, 0x61, 0x56, 0xb7, 0xe5, 0x99, 0x29, 0x9d,
	0x9f, 0x6b, 0x2c, 0x3f, 0xce, 0x3c, 0x41, 0x6f, 0xe0, 0xf9, 0x5a, 0xdc, 0x3f, 0xb9, 0xb2, 0xdf,
	0x91, 0xd3, 0xa9, 0x2a, 0x14, 0xf9, 0x77, 0x71, 0x0f, 0x8e, 0x7e, 0xef, 0xf2, 0xe8, 0xcf, 0x60,
	0x64, 0x81, 0x8d, 0x2e, 0xc6, 0xe8, 0x42, 0xea, 0x91, 0x63, 0x1e, 0x5e, 0x8c, 0xfe, 0xc5, 0xc5,
	0xb8, 0x86, 0x41, 0xc6, 0x72, 0x25, 0x78, 0x3a, 0xc0, 0x15, 0x87, 0xcc, 0x3b, 0x5f, 0x45, 0x91,
	0x6b, 0x21, 0x55, 0x7a, 0x85, 0x46, 0x3a, 0x4c, 0xd5, 0xc5, 0xb6, 0xec, 0x0d, 0xff, 0x7b, 0x3c,
	0xf6, 0xa4, 0xd9, 0x1f, 0x87, 0x05, 0x3e, 0xb4, 0x38, 0x08, 0x2d, 0x30, 0x94, 0x84, 0x86, 0x76,
	0x03, 0xfc, 0x5d, 0x7d, 0xf8, 0x13, 0x00, 0x00, 0xff, 0xff, 0x59, 0x00, 0xa1, 0x6d, 0xbf, 0x04,
	0x00, 0x00,
>>>>>>> feat: batch send block when synchronize
}
