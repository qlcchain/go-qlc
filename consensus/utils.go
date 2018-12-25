package consensus

import "github.com/qlcchain/go-qlc/p2p/protos"

func IsAckSignValidate(vote_a *protos.ConfirmAckBlock) bool {
	hash := vote_a.Blk.GetHash()
	verify := vote_a.Account.Verify(hash[:], vote_a.Signature[:])
	if verify != true {
		logger.Error("Ack Signature Verify Failed.")
	}
	return verify
}
