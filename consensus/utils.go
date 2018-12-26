package consensus

import "github.com/qlcchain/go-qlc/p2p/protos"

func IsAckSignValidate(va *protos.ConfirmAckBlock) bool {
	hash := va.Blk.GetHash()
	verify := va.Account.Verify(hash[:], va.Signature[:])
	if verify != true {
		logger.Error("Ack Signature Verify Failed.")
	}
	return verify
}
