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
	Fs []*types.FrontierBlock
}

// ToProto converts domain FrontierResponse into proto FrontierResponse
func FrontierResponseToProto(frs *FrontierResponse) ([]byte, error) {
	frData := make([][]byte, 0)
	for _, f := range frs.Fs {
		data, err := f.Serialize()
		if err != nil {
			return nil, err
		}
		frData = append(frData, data)
	}

	frPb := &pb.FrontierRsp{
		Frontiers: frData,
	}

	data, err := proto.Marshal(frPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FrontierResponseFromProto parse the data into frontier message
func FrontierResponseFromProto(data []byte) (*FrontierResponse, error) {
	bp := new(pb.FrontierRsp)
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}

	frontierResponse := &FrontierResponse{}

	for _, b := range bp.Frontiers {
		fb := &types.FrontierBlock{}

		if err := fb.Deserialize(b); err != nil {
			return nil, err
		}

		frontierResponse.Fs = append(frontierResponse.Fs, fb)
	}
	return frontierResponse, nil
}
