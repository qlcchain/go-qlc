package protos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos/pb"
)

type ConfirmAckBlock struct {
	Account   types.Address
	Signature types.Signature
	Sequence  uint32
	Hash      []types.Hash
}

// ToProto converts domain ConfirmAckBlock into proto ConfirmAckBlock
func ConfirmAckBlockToProto(confirmAck *ConfirmAckBlock) ([]byte, error) {
	bpPb := &pb.ConfirmAck{
		Account:   confirmAck.Account.Bytes(),
		Signature: confirmAck.Signature[:],
		Sequence:  confirmAck.Sequence,
	}

	for i, _ := range confirmAck.Hash {
		bpPb.Hash = append(bpPb.Hash, confirmAck.Hash[i][:])
	}

	data, err := proto.Marshal(bpPb)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ConfirmAckBlockFromProto parse the data into ConfirmAckBlock message
func ConfirmAckBlockFromProto(data []byte) (*ConfirmAckBlock, error) {
	ca := new(pb.ConfirmAck)
	if err := proto.Unmarshal(data, ca); err != nil {
		return nil, err
	}
	account, err := types.BytesToAddress(ca.Account)
	if err != nil {
		return nil, err
	}
	var sign types.Signature
	err = sign.UnmarshalBinary(ca.Signature)
	if err != nil {
		return nil, err
	}

	hash := make([]types.Hash, 0)
	for _, h := range ca.Hash {
		ha, err := types.BytesToHash(h)
		if err != nil {
			return nil, err
		}
		hash = append(hash, ha)
	}

	ack := &ConfirmAckBlock{
		Account:   account,
		Signature: sign,
		Sequence:  ca.Sequence,
		Hash:      hash,
	}
	return ack, nil
}
