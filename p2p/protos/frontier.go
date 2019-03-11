package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type FrontierReq struct {
	StartAddress types.Address
	Age          uint32
	Count        uint32
}

func NewFrontierReq(addr types.Address, Age, Count uint32) (packet *FrontierReq) {
	return &FrontierReq{
		StartAddress: addr,
		Age:          Age,
		Count:        Count,
	}
}

// ToProto converts domain frontier into proto frontier
func FrontierReqToProto(fr *FrontierReq) ([]byte, error) {
	//pb := new(pb.Frontier)
	address := fr.StartAddress.Bytes()
	frPb := &pb.FrontierReq{
		Address: address,
		Age:     fr.Age,
		Count:   fr.Count,
	}
	data, err := proto.Marshal(frPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FrontierReqFromProto parse the data into frontier message
func FrontierReqFromProto(data []byte) (*FrontierReq, error) {
	fr := new(pb.FrontierReq)

	if err := proto.Unmarshal(data, fr); err != nil {
		return nil, err
	}
	address, err := types.BytesToAddress(fr.Address)
	if err != nil {
		return nil, err
	}
	frq := &FrontierReq{
		StartAddress: address,
		Age:          fr.Age,
		Count:        fr.Count,
	}
	return frq, nil
}

type FrontierResponse struct {
	Frontier         *types.Frontier
	TotalFrontierNum uint32
}

func NewFrontierRsp(fr *types.Frontier, num uint32) (packet *FrontierResponse) {
	return &FrontierResponse{
		Frontier:         fr,
		TotalFrontierNum: num,
	}
}

// ToProto converts domain FrontierResponse into proto FrontierResponse
func FrontierResponseToProto(fr *FrontierResponse) ([]byte, error) {
	pi := &pb.FrontierRsp{
		TotalFrontierNum: fr.TotalFrontierNum,
		HeaderBlock:      fr.Frontier.HeaderBlock[:],
		OpenBlock:        fr.Frontier.OpenBlock[:],
	}
	data, err := proto.Marshal(pi)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FrontierResponseFromProto parse the data into frontier message
func FrontierResponseFromProto(data []byte) (*FrontierResponse, error) {
	fr := new(pb.FrontierRsp)
	frp := new(types.Frontier)
	if err := proto.Unmarshal(data, fr); err != nil {
		return nil, err
	}
	err := frp.HeaderBlock.UnmarshalBinary(fr.HeaderBlock[:])
	if err != nil {
		return nil, err
	}
	err = frp.OpenBlock.UnmarshalBinary(fr.OpenBlock[:])
	if err != nil {
		return nil, err
	}
	frPs := &FrontierResponse{
		TotalFrontierNum: fr.TotalFrontierNum,
		Frontier:         frp,
	}
	return frPs, nil
}
