// Code generated by protoc-gen-go. DO NOT EDIT.
// source: p2p/protos/pb/message.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
	return fileDescriptor_eb45e02bf5f44afc, []int{0}
}

func (m *FrontierReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FrontierReq.Unmarshal(m, b)
}
func (m *FrontierReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FrontierReq.Marshal(b, m, deterministic)
}
func (m *FrontierReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FrontierReq.Merge(m, src)
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
	TotalFrontierNum     uint32   `protobuf:"varint,1,opt,name=TotalFrontierNum,proto3" json:"TotalFrontierNum,omitempty"`
	HeaderBlock          []byte   `protobuf:"bytes,2,opt,name=HeaderBlock,proto3" json:"HeaderBlock,omitempty"`
	OpenBlock            []byte   `protobuf:"bytes,3,opt,name=OpenBlock,proto3" json:"OpenBlock,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FrontierRsp) Reset()         { *m = FrontierRsp{} }
func (m *FrontierRsp) String() string { return proto.CompactTextString(m) }
func (*FrontierRsp) ProtoMessage()    {}
func (*FrontierRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb45e02bf5f44afc, []int{1}
}

func (m *FrontierRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FrontierRsp.Unmarshal(m, b)
}
func (m *FrontierRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FrontierRsp.Marshal(b, m, deterministic)
}
func (m *FrontierRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FrontierRsp.Merge(m, src)
}
func (m *FrontierRsp) XXX_Size() int {
	return xxx_messageInfo_FrontierRsp.Size(m)
}
func (m *FrontierRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_FrontierRsp.DiscardUnknown(m)
}

var xxx_messageInfo_FrontierRsp proto.InternalMessageInfo

func (m *FrontierRsp) GetTotalFrontierNum() uint32 {
	if m != nil {
		return m.TotalFrontierNum
	}
	return 0
}

func (m *FrontierRsp) GetHeaderBlock() []byte {
	if m != nil {
		return m.HeaderBlock
	}
	return nil
}

func (m *FrontierRsp) GetOpenBlock() []byte {
	if m != nil {
		return m.OpenBlock
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
	return fileDescriptor_eb45e02bf5f44afc, []int{2}
}

func (m *BulkPullReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BulkPullReq.Unmarshal(m, b)
}
func (m *BulkPullReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BulkPullReq.Marshal(b, m, deterministic)
}
func (m *BulkPullReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BulkPullReq.Merge(m, src)
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
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	Blocks               []byte   `protobuf:"bytes,3,opt,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BulkPullRsp) Reset()         { *m = BulkPullRsp{} }
func (m *BulkPullRsp) String() string { return proto.CompactTextString(m) }
func (*BulkPullRsp) ProtoMessage()    {}
func (*BulkPullRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb45e02bf5f44afc, []int{3}
}

func (m *BulkPullRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BulkPullRsp.Unmarshal(m, b)
}
func (m *BulkPullRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BulkPullRsp.Marshal(b, m, deterministic)
}
func (m *BulkPullRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BulkPullRsp.Merge(m, src)
}
func (m *BulkPullRsp) XXX_Size() int {
	return xxx_messageInfo_BulkPullRsp.Size(m)
}
func (m *BulkPullRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_BulkPullRsp.DiscardUnknown(m)
}

var xxx_messageInfo_BulkPullRsp proto.InternalMessageInfo

func (m *BulkPullRsp) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

func (m *BulkPullRsp) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

func (m *BulkPullRsp) GetBlocks() []byte {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type BulkPushBlock struct {
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BulkPushBlock) Reset()         { *m = BulkPushBlock{} }
func (m *BulkPushBlock) String() string { return proto.CompactTextString(m) }
func (*BulkPushBlock) ProtoMessage()    {}
func (*BulkPushBlock) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb45e02bf5f44afc, []int{4}
}

func (m *BulkPushBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BulkPushBlock.Unmarshal(m, b)
}
func (m *BulkPushBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BulkPushBlock.Marshal(b, m, deterministic)
}
func (m *BulkPushBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BulkPushBlock.Merge(m, src)
}
func (m *BulkPushBlock) XXX_Size() int {
	return xxx_messageInfo_BulkPushBlock.Size(m)
}
func (m *BulkPushBlock) XXX_DiscardUnknown() {
	xxx_messageInfo_BulkPushBlock.DiscardUnknown(m)
}

var xxx_messageInfo_BulkPushBlock proto.InternalMessageInfo

func (m *BulkPushBlock) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

func (m *BulkPushBlock) GetBlock() []byte {
	if m != nil {
		return m.Block
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
	return fileDescriptor_eb45e02bf5f44afc, []int{5}
}

func (m *PublishBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PublishBlock.Unmarshal(m, b)
}
func (m *PublishBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PublishBlock.Marshal(b, m, deterministic)
}
func (m *PublishBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PublishBlock.Merge(m, src)
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
	Blocktype            uint32   `protobuf:"varint,1,opt,name=blocktype,proto3" json:"blocktype,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConfirmReq) Reset()         { *m = ConfirmReq{} }
func (m *ConfirmReq) String() string { return proto.CompactTextString(m) }
func (*ConfirmReq) ProtoMessage()    {}
func (*ConfirmReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_eb45e02bf5f44afc, []int{6}
}

func (m *ConfirmReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfirmReq.Unmarshal(m, b)
}
func (m *ConfirmReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfirmReq.Marshal(b, m, deterministic)
}
func (m *ConfirmReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfirmReq.Merge(m, src)
}
func (m *ConfirmReq) XXX_Size() int {
	return xxx_messageInfo_ConfirmReq.Size(m)
}
func (m *ConfirmReq) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfirmReq.DiscardUnknown(m)
}

var xxx_messageInfo_ConfirmReq proto.InternalMessageInfo

func (m *ConfirmReq) GetBlocktype() uint32 {
	if m != nil {
		return m.Blocktype
	}
	return 0
}

func (m *ConfirmReq) GetBlock() []byte {
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
	return fileDescriptor_eb45e02bf5f44afc, []int{7}
}

func (m *ConfirmAck) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfirmAck.Unmarshal(m, b)
}
func (m *ConfirmAck) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfirmAck.Marshal(b, m, deterministic)
}
func (m *ConfirmAck) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfirmAck.Merge(m, src)
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
	return fileDescriptor_eb45e02bf5f44afc, []int{8}
}

func (m *PovStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovStatus.Unmarshal(m, b)
}
func (m *PovStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovStatus.Marshal(b, m, deterministic)
}
func (m *PovStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovStatus.Merge(m, src)
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
	return fileDescriptor_eb45e02bf5f44afc, []int{9}
}

func (m *PovPublishBlock) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovPublishBlock.Unmarshal(m, b)
}
func (m *PovPublishBlock) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovPublishBlock.Marshal(b, m, deterministic)
}
func (m *PovPublishBlock) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovPublishBlock.Merge(m, src)
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
	return fileDescriptor_eb45e02bf5f44afc, []int{10}
}

func (m *PovPullBlockReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovPullBlockReq.Unmarshal(m, b)
}
func (m *PovPullBlockReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovPullBlockReq.Marshal(b, m, deterministic)
}
func (m *PovPullBlockReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovPullBlockReq.Merge(m, src)
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
	return fileDescriptor_eb45e02bf5f44afc, []int{11}
}

func (m *PovPullBlockRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PovPullBlockRsp.Unmarshal(m, b)
}
func (m *PovPullBlockRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PovPullBlockRsp.Marshal(b, m, deterministic)
}
func (m *PovPullBlockRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PovPullBlockRsp.Merge(m, src)
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

func init() { proto.RegisterFile("p2p/protos/pb/message.proto", fileDescriptor_eb45e02bf5f44afc) }

var fileDescriptor_eb45e02bf5f44afc = []byte{
	// 541 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x96, 0xeb, 0x24, 0x25, 0x93, 0x44, 0x54, 0x2b, 0x54, 0x59, 0xd0, 0x43, 0x64, 0x71, 0x88,
	0x38, 0x10, 0x09, 0x5e, 0x80, 0x24, 0x14, 0x72, 0x40, 0x34, 0x72, 0x72, 0xe1, 0x68, 0x3b, 0x43,
	0x62, 0xc5, 0xf1, 0xba, 0xbb, 0xeb, 0x4a, 0x7d, 0x0a, 0x9e, 0x85, 0x87, 0xe1, 0x7d, 0xd0, 0xce,
	0xae, 0xed, 0x6d, 0x10, 0x82, 0xaa, 0xb7, 0xfd, 0xbe, 0xf1, 0x7e, 0x3b, 0xdf, 0xfc, 0x18, 0x5e,
	0x95, 0xef, 0xca, 0x69, 0x29, 0xb8, 0xe2, 0x72, 0x5a, 0x26, 0xd3, 0x23, 0x4a, 0x19, 0xef, 0xf0,
	0x2d, 0x31, 0xec, 0xac, 0x4c, 0xc2, 0x1b, 0x18, 0x7c, 0x12, 0xbc, 0x50, 0x19, 0x8a, 0x08, 0x6f,
	0x59, 0x00, 0xe7, 0xb3, 0xed, 0x56, 0xa0, 0x94, 0x81, 0x37, 0xf6, 0x26, 0xc3, 0xa8, 0x86, 0xec,
	0x02, 0xfc, 0xd9, 0x0e, 0x83, 0xb3, 0xb1, 0x37, 0x19, 0x45, 0xfa, 0xc8, 0x5e, 0x40, 0x77, 0xc1,
	0xab, 0x42, 0x05, 0x3e, 0x71, 0x06, 0x84, 0xf7, 0x8e, 0xa0, 0x2c, 0xd9, 0x1b, 0xb8, 0xd8, 0x70,
	0x15, 0xe7, 0x35, 0xf7, 0xb5, 0x3a, 0x92, 0xf2, 0x28, 0xfa, 0x83, 0x67, 0x63, 0x18, 0x2c, 0x31,
	0xde, 0xa2, 0x98, 0xe7, 0x3c, 0x3d, 0xd0, 0x53, 0xc3, 0xc8, 0xa5, 0xd8, 0x15, 0xf4, 0x6f, 0x4a,
	0x2c, 0x4c, 0xdc, 0xa7, 0x78, 0x4b, 0x84, 0x3f, 0x3c, 0x18, 0xcc, 0xab, 0xfc, 0xb0, 0xaa, 0xf2,
	0x5c, 0x9b, 0xb9, 0x82, 0xfe, 0x5a, 0xc5, 0x42, 0x2d, 0x63, 0xb9, 0xb7, 0x76, 0x5a, 0x42, 0x5b,
	0xbd, 0x2e, 0xb6, 0x14, 0x33, 0x2f, 0xd5, 0x90, 0xbd, 0x84, 0x67, 0x5a, 0x62, 0x73, 0x5f, 0xa2,
	0xf5, 0xd6, 0xe0, 0xd6, 0x74, 0xc7, 0x31, 0xcd, 0x2e, 0xa1, 0xa7, 0x6f, 0xa2, 0x0c, 0xba, 0x24,
	0x65, 0x51, 0xf8, 0xcd, 0x49, 0x48, 0x96, 0x3a, 0xa1, 0x44, 0x67, 0xaa, 0xb4, 0xb2, 0xa9, 0x42,
	0x4b, 0x68, 0xe9, 0xc4, 0x31, 0x6e, 0x80, 0x96, 0xa6, 0x83, 0xb4, 0x7e, 0x2d, 0x0a, 0x17, 0x30,
	0x32, 0xd2, 0x72, 0xdf, 0xd4, 0xe6, 0xb1, 0xe2, 0xe1, 0x1c, 0x86, 0xab, 0x2a, 0xc9, 0xb3, 0xa7,
	0x68, 0x7c, 0x00, 0x58, 0xf0, 0xe2, 0x7b, 0x26, 0x8e, 0xb6, 0xe6, 0x8f, 0x56, 0x50, 0x8d, 0xc2,
	0x2c, 0x3d, 0xd0, 0x08, 0xa6, 0x29, 0xd5, 0xb8, 0x1e, 0x41, 0x03, 0xa9, 0x9f, 0xd9, 0xae, 0x88,
	0x55, 0x25, 0xd0, 0x2a, 0xb4, 0x84, 0xee, 0xda, 0x1a, 0x6f, 0x2b, 0x2c, 0xd2, 0xa6, 0x6b, 0x35,
	0x66, 0x0c, 0x3a, 0xd4, 0xe8, 0xce, 0xd8, 0x9f, 0x0c, 0x23, 0x3a, 0x87, 0x3f, 0x3d, 0xe8, 0xaf,
	0xf8, 0xdd, 0x5a, 0xc5, 0xaa, 0x92, 0xec, 0x35, 0x8c, 0x16, 0x95, 0x10, 0x58, 0xa8, 0x25, 0x66,
	0xbb, 0xbd, 0x79, 0xbb, 0x13, 0x3d, 0x24, 0xf5, 0x84, 0xd6, 0x44, 0x3b, 0x37, 0x2e, 0xa5, 0xbf,
	0xf8, 0x8c, 0x05, 0xca, 0x4c, 0xd2, 0x17, 0xa6, 0x67, 0x2e, 0xa5, 0x5d, 0xd8, 0x0b, 0x9b, 0x8f,
	0x34, 0x45, 0xc3, 0xa8, 0x25, 0x74, 0x74, 0x93, 0x1d, 0x51, 0xaa, 0xf8, 0x58, 0xd2, 0x30, 0xf9,
	0x51, 0x4b, 0x84, 0xd7, 0xf0, 0x7c, 0xc5, 0xef, 0x9e, 0xdc, 0xb2, 0x5f, 0x9e, 0xd5, 0xc9, 0x73,
	0x12, 0xf9, 0x77, 0xe3, 0x1e, 0xac, 0xd2, 0xd9, 0xe9, 0x2a, 0x8d, 0x61, 0x60, 0x80, 0x29, 0x9d,
	0x4f, 0xa5, 0x73, 0xa9, 0xbf, 0xac, 0x8d, 0xbb, 0x68, 0xdd, 0x93, 0x45, 0xbb, 0x84, 0x5e, 0x84,
	0xb1, 0xe4, 0x45, 0xd0, 0xa3, 0x88, 0x45, 0xfa, 0xce, 0x17, 0x9e, 0xc6, 0x8a, 0x0b, 0x19, 0x9c,
	0x53, 0x22, 0x0d, 0x0e, 0xe5, 0x89, 0xad, 0xff, 0x59, 0x39, 0x33, 0x69, 0xe6, 0xb7, 0x66, 0x40,
	0x5b, 0x34, 0xff, 0x64, 0x11, 0x6d, 0x42, 0x1d, 0x37, 0xa1, 0xa4, 0x47, 0x3f, 0xd3, 0xf7, 0xbf,
	0x03, 0x00, 0x00, 0xff, 0xff, 0xee, 0x78, 0xd8, 0x5b, 0x6b, 0x05, 0x00, 0x00,
}
