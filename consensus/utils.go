package consensus

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

type MsgType byte

const (
	MsgPublishReq MsgType = iota
	MsgConfirmReq
	MsgConfirmAck
	MsgSync
	MsgGenerateBlock
)

type BlockSource struct {
	Block     *types.StateBlock
	BlockFrom types.SynchronizedKind
	Type      MsgType
	Para      interface{}
	MsgFrom   string
}

func IsAckSignValidate(va *protos.ConfirmAckBlock) bool {
	hash := va.Hash
	verify := va.Account.Verify(hash[:], va.Signature[:])
	return verify
}
