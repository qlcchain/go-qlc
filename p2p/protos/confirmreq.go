package protos

import (
	"github.com/gogo/protobuf/proto"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type ConfirmReqBlock struct {
	Blk []*types.StateBlock
}

// ToProto converts domain ConfirmReqBlock into proto ConfirmReqBlock
func ConfirmReqBlockToProto(confirmReq *ConfirmReqBlock) ([]byte, error) {
	blkData := make([][]byte, 0)

	for _, blk := range confirmReq.Blk {
		data, err := blk.Serialize()
		if err != nil {
			return nil, err
		}
		blkData = append(blkData, data)
	}

	bpPb := &pb.ConfirmReq{
		Block: blkData,
	}

	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ConfirmReqBlockFromProto parse the data into ConfirmReqBlock message
func ConfirmReqBlockFromProto(data []byte) (*ConfirmReqBlock, error) {
	bp := new(pb.ConfirmReq)
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}

	confirmReqBlock := &ConfirmReqBlock{}

	for _, b := range bp.Block {
		blk := &types.StateBlock{}

		if err := blk.Deserialize(b); err != nil {
			return nil, err
		}

		confirmReqBlock.Blk = append(confirmReqBlock.Blk, blk)
	}

	return confirmReqBlock, nil
}
