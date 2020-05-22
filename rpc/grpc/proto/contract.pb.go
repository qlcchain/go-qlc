// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0
// 	protoc        v3.7.1
// source: contract.proto

package proto

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
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

type PackContractDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AbiStr     string `protobuf:"bytes,1,opt,name=abiStr,proto3" json:"abiStr,omitempty"`
	MethodName string `protobuf:"bytes,2,opt,name=methodName,proto3" json:"methodName,omitempty"`
	Params     []byte `protobuf:"bytes,3,opt,name=params,proto3" json:"params,omitempty"`
}

func (x *PackContractDataRequest) Reset() {
	*x = PackContractDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PackContractDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackContractDataRequest) ProtoMessage() {}

func (x *PackContractDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackContractDataRequest.ProtoReflect.Descriptor instead.
func (*PackContractDataRequest) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{0}
}

func (x *PackContractDataRequest) GetAbiStr() string {
	if x != nil {
		return x.AbiStr
	}
	return ""
}

func (x *PackContractDataRequest) GetMethodName() string {
	if x != nil {
		return x.MethodName
	}
	return ""
}

func (x *PackContractDataRequest) GetParams() []byte {
	if x != nil {
		return x.Params
	}
	return nil
}

type PackChainContractDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContractAddress string `protobuf:"bytes,1,opt,name=contractAddress,proto3" json:"contractAddress,omitempty"`
	MethodName      string `protobuf:"bytes,2,opt,name=methodName,proto3" json:"methodName,omitempty"`
	Params          []byte `protobuf:"bytes,3,opt,name=params,proto3" json:"params,omitempty"`
}

func (x *PackChainContractDataRequest) Reset() {
	*x = PackChainContractDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PackChainContractDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PackChainContractDataRequest) ProtoMessage() {}

func (x *PackChainContractDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PackChainContractDataRequest.ProtoReflect.Descriptor instead.
func (*PackChainContractDataRequest) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{1}
}

func (x *PackChainContractDataRequest) GetContractAddress() string {
	if x != nil {
		return x.ContractAddress
	}
	return ""
}

func (x *PackChainContractDataRequest) GetMethodName() string {
	if x != nil {
		return x.MethodName
	}
	return ""
}

func (x *PackChainContractDataRequest) GetParams() []byte {
	if x != nil {
		return x.Params
	}
	return nil
}

type ContractSendBlockPara struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address     string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	TokenName   string   `protobuf:"bytes,2,opt,name=tokenName,proto3" json:"tokenName,omitempty"`
	To          string   `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Amount      int64    `protobuf:"varint,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Data        string   `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
	PrivateFrom string   `protobuf:"bytes,6,opt,name=privateFrom,proto3" json:"privateFrom,omitempty"`
	PrivateFor  []string `protobuf:"bytes,7,rep,name=privateFor,proto3" json:"privateFor,omitempty"`
	String_     string   `protobuf:"bytes,8,opt,name=string,proto3" json:"string,omitempty"`
	EnclaveKey  string   `protobuf:"bytes,9,opt,name=enclaveKey,proto3" json:"enclaveKey,omitempty"`
}

func (x *ContractSendBlockPara) Reset() {
	*x = ContractSendBlockPara{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContractSendBlockPara) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContractSendBlockPara) ProtoMessage() {}

func (x *ContractSendBlockPara) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContractSendBlockPara.ProtoReflect.Descriptor instead.
func (*ContractSendBlockPara) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{2}
}

func (x *ContractSendBlockPara) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *ContractSendBlockPara) GetTokenName() string {
	if x != nil {
		return x.TokenName
	}
	return ""
}

func (x *ContractSendBlockPara) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *ContractSendBlockPara) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *ContractSendBlockPara) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

func (x *ContractSendBlockPara) GetPrivateFrom() string {
	if x != nil {
		return x.PrivateFrom
	}
	return ""
}

func (x *ContractSendBlockPara) GetPrivateFor() []string {
	if x != nil {
		return x.PrivateFor
	}
	return nil
}

func (x *ContractSendBlockPara) GetString_() string {
	if x != nil {
		return x.String_
	}
	return ""
}

func (x *ContractSendBlockPara) GetEnclaveKey() string {
	if x != nil {
		return x.EnclaveKey
	}
	return ""
}

type ContractRewardBlockPara struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SendHash string `protobuf:"bytes,1,opt,name=sendHash,proto3" json:"sendHash,omitempty"`
}

func (x *ContractRewardBlockPara) Reset() {
	*x = ContractRewardBlockPara{}
	if protoimpl.UnsafeEnabled {
		mi := &file_contract_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContractRewardBlockPara) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContractRewardBlockPara) ProtoMessage() {}

func (x *ContractRewardBlockPara) ProtoReflect() protoreflect.Message {
	mi := &file_contract_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContractRewardBlockPara.ProtoReflect.Descriptor instead.
func (*ContractRewardBlockPara) Descriptor() ([]byte, []int) {
	return file_contract_proto_rawDescGZIP(), []int{3}
}

func (x *ContractRewardBlockPara) GetSendHash() string {
	if x != nil {
		return x.SendHash
	}
	return ""
}

var File_contract_proto protoreflect.FileDescriptor

var file_contract_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x62, 0x61, 0x73, 0x69, 0x63, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x69, 0x0a, 0x17, 0x50, 0x61, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x61, 0x62, 0x69, 0x53, 0x74, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x61, 0x62, 0x69, 0x53, 0x74, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x65, 0x74, 0x68,
	0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x22, 0x80,
	0x01, 0x0a, 0x1c, 0x50, 0x61, 0x63, 0x6b, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x28, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x6d, 0x65, 0x74,
	0x68, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x61, 0x72,
	0x61, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x61, 0x72, 0x61, 0x6d,
	0x73, 0x22, 0x85, 0x02, 0x0a, 0x15, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x53, 0x65,
	0x6e, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x50, 0x61, 0x72, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x74, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x20, 0x0a, 0x0b, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x46, 0x72, 0x6f, 0x6d, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x46, 0x72, 0x6f,
	0x6d, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x46, 0x6f, 0x72, 0x18,
	0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x46, 0x6f,
	0x72, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x6e, 0x63,
	0x6c, 0x61, 0x76, 0x65, 0x4b, 0x65, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65,
	0x6e, 0x63, 0x6c, 0x61, 0x76, 0x65, 0x4b, 0x65, 0x79, 0x22, 0x35, 0x0a, 0x17, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x61, 0x63, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x50, 0x61, 0x72, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x6e, 0x64, 0x48, 0x61, 0x73, 0x68,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x6e, 0x64, 0x48, 0x61, 0x73, 0x68,
	0x32, 0x9d, 0x05, 0x0a, 0x0b, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x50, 0x49,
	0x12, 0x63, 0x0a, 0x17, 0x47, 0x65, 0x74, 0x41, 0x62, 0x69, 0x42, 0x79, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x0e, 0x2e, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x1a, 0x0d, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x23, 0x12, 0x21, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x67, 0x65,
	0x74, 0x41, 0x62, 0x69, 0x42, 0x79, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x67, 0x0a, 0x10, 0x50, 0x61, 0x63, 0x6b, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x22, 0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x22,
	0x1a, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x70, 0x61, 0x63, 0x6b, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x3a, 0x01, 0x2a, 0x12, 0x76,
	0x0a, 0x15, 0x50, 0x61, 0x63, 0x6b, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x61, 0x63, 0x74, 0x44, 0x61, 0x74, 0x61, 0x12, 0x23, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x50, 0x61, 0x63, 0x6b, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63,
	0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x22, 0x2a, 0x82, 0xd3, 0xe4, 0x93,
	0x02, 0x24, 0x22, 0x1f, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x70, 0x61,
	0x63, 0x6b, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x3a, 0x01, 0x2a, 0x12, 0x66, 0x0a, 0x13, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x10, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x22, 0x25, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x12,
	0x1d, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72,
	0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x6c,
	0x0a, 0x11, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x53, 0x65, 0x6e, 0x64, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x12, 0x1c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x61, 0x63, 0x74, 0x53, 0x65, 0x6e, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x50, 0x61, 0x72,
	0x61, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x42,
	0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x26, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x20, 0x22, 0x1b, 0x2f, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65,
	0x53, 0x65, 0x6e, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x3a, 0x01, 0x2a, 0x12, 0x72, 0x0a, 0x13,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x12, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x61, 0x63, 0x74, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x50,
	0x61, 0x72, 0x61, 0x1a, 0x11, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x22, 0x28, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x22, 0x22, 0x1d,
	0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x77, 0x61, 0x72, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x3a, 0x01, 0x2a,
	0x42, 0x09, 0x5a, 0x07, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_contract_proto_rawDescOnce sync.Once
	file_contract_proto_rawDescData = file_contract_proto_rawDesc
)

func file_contract_proto_rawDescGZIP() []byte {
	file_contract_proto_rawDescOnce.Do(func() {
		file_contract_proto_rawDescData = protoimpl.X.CompressGZIP(file_contract_proto_rawDescData)
	})
	return file_contract_proto_rawDescData
}

var file_contract_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_contract_proto_goTypes = []interface{}{
	(*PackContractDataRequest)(nil),      // 0: proto.PackContractDataRequest
	(*PackChainContractDataRequest)(nil), // 1: proto.PackChainContractDataRequest
	(*ContractSendBlockPara)(nil),        // 2: proto.ContractSendBlockPara
	(*ContractRewardBlockPara)(nil),      // 3: proto.ContractRewardBlockPara
	(*Address)(nil),                      // 4: types.Address
	(*empty.Empty)(nil),                  // 5: google.protobuf.Empty
	(*String)(nil),                       // 6: proto.String
	(*Bytes)(nil),                        // 7: proto.Bytes
	(*Addresses)(nil),                    // 8: types.Addresses
	(*StateBlock)(nil),                   // 9: types.StateBlock
}
var file_contract_proto_depIdxs = []int32{
	4, // 0: proto.ContractAPI.GetAbiByContractAddress:input_type -> types.Address
	0, // 1: proto.ContractAPI.PackContractData:input_type -> proto.PackContractDataRequest
	1, // 2: proto.ContractAPI.PackChainContractData:input_type -> proto.PackChainContractDataRequest
	5, // 3: proto.ContractAPI.ContractAddressList:input_type -> google.protobuf.Empty
	2, // 4: proto.ContractAPI.GenerateSendBlock:input_type -> proto.ContractSendBlockPara
	3, // 5: proto.ContractAPI.GenerateRewardBlock:input_type -> proto.ContractRewardBlockPara
	6, // 6: proto.ContractAPI.GetAbiByContractAddress:output_type -> proto.String
	7, // 7: proto.ContractAPI.PackContractData:output_type -> proto.Bytes
	7, // 8: proto.ContractAPI.PackChainContractData:output_type -> proto.Bytes
	8, // 9: proto.ContractAPI.ContractAddressList:output_type -> types.Addresses
	9, // 10: proto.ContractAPI.GenerateSendBlock:output_type -> types.StateBlock
	9, // 11: proto.ContractAPI.GenerateRewardBlock:output_type -> types.StateBlock
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_contract_proto_init() }
func file_contract_proto_init() {
	if File_contract_proto != nil {
		return
	}
	file_types_basic_proto_init()
	file_common_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_contract_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PackContractDataRequest); i {
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
		file_contract_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PackChainContractDataRequest); i {
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
		file_contract_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContractSendBlockPara); i {
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
		file_contract_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContractRewardBlockPara); i {
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
			RawDescriptor: file_contract_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_contract_proto_goTypes,
		DependencyIndexes: file_contract_proto_depIdxs,
		MessageInfos:      file_contract_proto_msgTypes,
	}.Build()
	File_contract_proto = out.File
	file_contract_proto_rawDesc = nil
	file_contract_proto_goTypes = nil
	file_contract_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ContractAPIClient is the client API for ContractAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ContractAPIClient interface {
	GetAbiByContractAddress(ctx context.Context, in *Address, opts ...grpc.CallOption) (*String, error)
	PackContractData(ctx context.Context, in *PackContractDataRequest, opts ...grpc.CallOption) (*Bytes, error)
	PackChainContractData(ctx context.Context, in *PackChainContractDataRequest, opts ...grpc.CallOption) (*Bytes, error)
	ContractAddressList(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Addresses, error)
	GenerateSendBlock(ctx context.Context, in *ContractSendBlockPara, opts ...grpc.CallOption) (*StateBlock, error)
	GenerateRewardBlock(ctx context.Context, in *ContractRewardBlockPara, opts ...grpc.CallOption) (*StateBlock, error)
}

type contractAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewContractAPIClient(cc grpc.ClientConnInterface) ContractAPIClient {
	return &contractAPIClient{cc}
}

func (c *contractAPIClient) GetAbiByContractAddress(ctx context.Context, in *Address, opts ...grpc.CallOption) (*String, error) {
	out := new(String)
	err := c.cc.Invoke(ctx, "/proto.ContractAPI/GetAbiByContractAddress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractAPIClient) PackContractData(ctx context.Context, in *PackContractDataRequest, opts ...grpc.CallOption) (*Bytes, error) {
	out := new(Bytes)
	err := c.cc.Invoke(ctx, "/proto.ContractAPI/PackContractData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractAPIClient) PackChainContractData(ctx context.Context, in *PackChainContractDataRequest, opts ...grpc.CallOption) (*Bytes, error) {
	out := new(Bytes)
	err := c.cc.Invoke(ctx, "/proto.ContractAPI/PackChainContractData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractAPIClient) ContractAddressList(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Addresses, error) {
	out := new(Addresses)
	err := c.cc.Invoke(ctx, "/proto.ContractAPI/ContractAddressList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractAPIClient) GenerateSendBlock(ctx context.Context, in *ContractSendBlockPara, opts ...grpc.CallOption) (*StateBlock, error) {
	out := new(StateBlock)
	err := c.cc.Invoke(ctx, "/proto.ContractAPI/GenerateSendBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractAPIClient) GenerateRewardBlock(ctx context.Context, in *ContractRewardBlockPara, opts ...grpc.CallOption) (*StateBlock, error) {
	out := new(StateBlock)
	err := c.cc.Invoke(ctx, "/proto.ContractAPI/GenerateRewardBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ContractAPIServer is the server API for ContractAPI service.
type ContractAPIServer interface {
	GetAbiByContractAddress(context.Context, *Address) (*String, error)
	PackContractData(context.Context, *PackContractDataRequest) (*Bytes, error)
	PackChainContractData(context.Context, *PackChainContractDataRequest) (*Bytes, error)
	ContractAddressList(context.Context, *empty.Empty) (*Addresses, error)
	GenerateSendBlock(context.Context, *ContractSendBlockPara) (*StateBlock, error)
	GenerateRewardBlock(context.Context, *ContractRewardBlockPara) (*StateBlock, error)
}

// UnimplementedContractAPIServer can be embedded to have forward compatible implementations.
type UnimplementedContractAPIServer struct {
}

func (*UnimplementedContractAPIServer) GetAbiByContractAddress(context.Context, *Address) (*String, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAbiByContractAddress not implemented")
}
func (*UnimplementedContractAPIServer) PackContractData(context.Context, *PackContractDataRequest) (*Bytes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PackContractData not implemented")
}
func (*UnimplementedContractAPIServer) PackChainContractData(context.Context, *PackChainContractDataRequest) (*Bytes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PackChainContractData not implemented")
}
func (*UnimplementedContractAPIServer) ContractAddressList(context.Context, *empty.Empty) (*Addresses, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ContractAddressList not implemented")
}
func (*UnimplementedContractAPIServer) GenerateSendBlock(context.Context, *ContractSendBlockPara) (*StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateSendBlock not implemented")
}
func (*UnimplementedContractAPIServer) GenerateRewardBlock(context.Context, *ContractRewardBlockPara) (*StateBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateRewardBlock not implemented")
}

func RegisterContractAPIServer(s *grpc.Server, srv ContractAPIServer) {
	s.RegisterService(&_ContractAPI_serviceDesc, srv)
}

func _ContractAPI_GetAbiByContractAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Address)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractAPIServer).GetAbiByContractAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ContractAPI/GetAbiByContractAddress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractAPIServer).GetAbiByContractAddress(ctx, req.(*Address))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractAPI_PackContractData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PackContractDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractAPIServer).PackContractData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ContractAPI/PackContractData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractAPIServer).PackContractData(ctx, req.(*PackContractDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractAPI_PackChainContractData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PackChainContractDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractAPIServer).PackChainContractData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ContractAPI/PackChainContractData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractAPIServer).PackChainContractData(ctx, req.(*PackChainContractDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractAPI_ContractAddressList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractAPIServer).ContractAddressList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ContractAPI/ContractAddressList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractAPIServer).ContractAddressList(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractAPI_GenerateSendBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContractSendBlockPara)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractAPIServer).GenerateSendBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ContractAPI/GenerateSendBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractAPIServer).GenerateSendBlock(ctx, req.(*ContractSendBlockPara))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractAPI_GenerateRewardBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContractRewardBlockPara)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractAPIServer).GenerateRewardBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.ContractAPI/GenerateRewardBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractAPIServer).GenerateRewardBlock(ctx, req.(*ContractRewardBlockPara))
	}
	return interceptor(ctx, in, info, handler)
}

var _ContractAPI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ContractAPI",
	HandlerType: (*ContractAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAbiByContractAddress",
			Handler:    _ContractAPI_GetAbiByContractAddress_Handler,
		},
		{
			MethodName: "PackContractData",
			Handler:    _ContractAPI_PackContractData_Handler,
		},
		{
			MethodName: "PackChainContractData",
			Handler:    _ContractAPI_PackChainContractData_Handler,
		},
		{
			MethodName: "ContractAddressList",
			Handler:    _ContractAPI_ContractAddressList_Handler,
		},
		{
			MethodName: "GenerateSendBlock",
			Handler:    _ContractAPI_GenerateSendBlock_Handler,
		},
		{
			MethodName: "GenerateRewardBlock",
			Handler:    _ContractAPI_GenerateRewardBlock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contract.proto",
}
