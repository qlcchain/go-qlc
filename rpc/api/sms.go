package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"
)

type SMSApi struct {
	ledger *ledger.Ledger
	logger *zap.SugaredLogger
}

func NewSMSApi(ledger *ledger.Ledger) *SMSApi {
	return &SMSApi{ledger: ledger, logger: log.NewLogger("api_sms")}
}

func (s *SMSApi) getSenderOrReceiver(hashes []types.Hash, count int, offset *int) ([]*APIBlock, error) {
	c, o, err := checkOffset(count, offset)
	if err != nil {
		return nil, err
	}

	var hs []types.Hash
	if len(hashes) > o {
		if len(hashes) >= o+c {
			hs = hashes[o : c+o]
		} else {
			hs = hashes[o:]
		}
	} else {
		return make([]*APIBlock, 0), nil
	}

	ab := make([]*APIBlock, 0)
	for _, h := range hs {
		block, err := s.ledger.GetStateBlock(h)
		if err != nil {
			return nil, err
		}
		b, err := generateAPIBlock(s.ledger, block)
		if err != nil {
			return nil, err
		}
		ab = append(ab, b)
	}
	return ab, nil

}

func (s *SMSApi) SenderBlocks(sender string, count int, offset *int) ([]*APIBlock, error) {
	hashes, err := s.ledger.GetSenderBlocks(sender)
	if err != nil {
		return nil, err
	}
	return s.getSenderOrReceiver(hashes, count, offset)
}

func (s *SMSApi) ReceiverBlocks(receiver string, count int, offset *int) ([]*APIBlock, error) {
	hashes, err := s.ledger.GetReceiverBlocks(receiver)
	if err != nil {
		return nil, err
	}
	return s.getSenderOrReceiver(hashes, count, offset)
}

func (s *SMSApi) SenderBlocksCount(sender string) (int, error) {
	hashes, err := s.ledger.GetSenderBlocks(sender)
	if err != nil {
		return 0, err
	}
	var num int
	num = len(hashes)
	return num, nil
}

func (s *SMSApi) ReceiverBlocksCount(receiver string) (int, error) {
	hashes, err := s.ledger.GetReceiverBlocks(receiver)
	if err != nil {
		return 0, err
	}
	var num int
	num = len(hashes)
	return num, nil
}

func (s *SMSApi) MessageHash(message string) (types.Hash, error) {
	m := util.String2Bytes(message)
	hash, err := types.HashBytes(m)
	if err != nil {
		return types.ZeroHash, err
	}
	return hash, nil
}

func (s *SMSApi) MessageBlock(hash types.Hash) (*APIBlock, error) {
	block, err := s.ledger.GetMessageBlock(hash)
	if err != nil {
		return nil, err
	}
	b, err := generateAPIBlock(s.ledger, block)
	if err != nil {
		return nil, err
	}
	return b, nil
}
