package protos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

const (
	PullTypeSegment = iota
	PullTypeBackward
	PullTypeForward
	PullTypeBatch
)

type BulkPullReqPacket struct {
	StartHash types.Hash
	EndHash   types.Hash
	PullType  uint32
	Count     uint32
	Hashes    []*types.Hash
}

func NewBulkPullReqPacket(start, end types.Hash) (packet *BulkPullReqPacket) {
	return &BulkPullReqPacket{
		StartHash: start,
		EndHash:   end,
	}
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPullReqPacketToProto(bp *BulkPullReqPacket) ([]byte, error) {
	hashBytes := make([]byte, 0)
	for _, locItem := range bp.Hashes {
		hashBytes = append(hashBytes, locItem.Bytes()...)
	}

	bpPb := &pb.BulkPullReq{
		StartHash: bp.StartHash[:],
		EndHash:   bp.EndHash[:],
		PullType:  bp.PullType,
		Count:     bp.Count,
		Hashes:    hashBytes,
	}
	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPullReqPacketFromProto(data []byte) (*BulkPullReqPacket, error) {
	bp := new(pb.BulkPullReq)
	var start, end types.Hash
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}
	err := start.UnmarshalBinary(bp.StartHash)
	if err != nil {
		return nil, err
	}
	err = end.UnmarshalBinary(bp.EndHash)
	if err != nil {
		return nil, err
	}

	hashes := make([]*types.Hash, 0)
	if len(bp.Hashes) > 0 {
		if len(bp.Hashes)%types.HashSize != 0 {
			return nil, fmt.Errorf("invalid hashes field length %d", len(bp.Hashes))
		}
		for hashIdx := 0; hashIdx < len(bp.Hashes); hashIdx += types.HashSize {
			hash, err := types.BytesToHash(bp.Hashes[hashIdx : hashIdx+types.HashSize])
			if err != nil {
				return nil, err
			}
			hashes = append(hashes, &hash)
		}
	}

	bpRp := &BulkPullReqPacket{
		StartHash: start,
		EndHash:   end,
		PullType:  bp.PullType,
		Count:     bp.Count,
		Hashes:    hashes,
	}
	return bpRp, nil
}

type BulkPullRspPacket struct {
	Blocks types.StateBlockList
}

// ToProto converts domain BulkPull into proto BulkPull
func BulkPullRspPacketToProto(bp *BulkPullRspPacket) ([]byte, error) {
	bpPb := &pb.BulkPullRsp{}

	if len(bp.Blocks) > 0 {
		blockBytes, err := bp.Blocks.Serialize()
		if err != nil {
			return nil, err
		}
		bpPb.Blocks = blockBytes
	}

	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// BulkPullPacketFromProto parse the data into BulkPull message
func BulkPullRspPacketFromProto(data []byte) (*BulkPullRspPacket, error) {
	bp := new(pb.BulkPullRsp)
	if err := proto.Unmarshal(data, bp); err != nil {
		return nil, err
	}

	bpRp := &BulkPullRspPacket{}

	if len(bp.Blocks) > 0 {
		blocks := make(types.StateBlockList, 0)
		err := blocks.Deserialize(bp.Blocks)
		if err != nil {
			return nil, err
		}
		bpRp.Blocks = blocks
	}
	return bpRp, nil
}
