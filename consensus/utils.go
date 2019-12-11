package consensus

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/process"
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
	Gap       process.ProcessResult
	GapHash   types.Hash
}

func IsAckSignValidate(va *protos.ConfirmAckBlock) bool {
	hashBytes := make([]byte, 0)
	for _, h := range va.Hash {
		hashBytes = append(hashBytes, h[:]...)
	}
	signHash, _ := types.HashBytes(hashBytes)

	verify := va.Account.Verify(signHash[:], va.Signature[:])
	return verify
}
