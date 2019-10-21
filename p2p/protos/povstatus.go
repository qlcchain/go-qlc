package protos

import (
	"github.com/gogo/protobuf/proto"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type PovStatus struct {
	CurrentHeight uint64
	CurrentTD     []byte
	CurrentHash   types.Hash
	GenesisHash   types.Hash
	Timestamp     int64
}

func PovStatusToProto(status *PovStatus) ([]byte, error) {
	pbStatus := &pb.PovStatus{
		CurrentHeight: status.CurrentHeight,
		CurrentTD:     status.CurrentTD,
		CurrentHash:   status.CurrentHash.Bytes(),
		GenesisHash:   status.GenesisHash.Bytes(),
		Timestamp:     status.Timestamp,
	}
	data, err := proto.Marshal(pbStatus)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func PovStatusFromProto(data []byte) (*PovStatus, error) {
	pbStatus := new(pb.PovStatus)
	if err := proto.Unmarshal(data, pbStatus); err != nil {
		return nil, err
	}
	status := new(PovStatus)
	status.Timestamp = pbStatus.Timestamp
	status.CurrentHeight = pbStatus.CurrentHeight
	status.CurrentTD = pbStatus.CurrentTD
	err := status.CurrentHash.UnmarshalBinary(pbStatus.CurrentHash)
	if err != nil {
		return nil, err
	}
	err = status.GenesisHash.UnmarshalBinary(pbStatus.GenesisHash)
	if err != nil {
		return nil, err
	}
	return status, nil
}
