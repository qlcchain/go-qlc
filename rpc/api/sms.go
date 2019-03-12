package api

import (
	"encoding/json"

	"github.com/qlcchain/go-qlc/common/types"
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

func (s *SMSApi) getSenderOrReceiver(hashes []types.Hash) ([]*APIBlock, error) {
	ab := make([]*APIBlock, 0)
	for _, h := range hashes {
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

func (s *SMSApi) PhoneBlocks(sender string) (map[string][]*APIBlock, error) {
	sHash, err := s.ledger.GetSenderBlocks(sender)
	if err != nil {
		return nil, err
	}
	senders, err := s.getSenderOrReceiver(sHash)
	if err != nil {
		return nil, err
	}
	rHash, err := s.ledger.GetReceiverBlocks(sender)
	if err != nil {
		return nil, err
	}
	receivers, err := s.getSenderOrReceiver(rHash)
	if err != nil {
		return nil, err
	}
	abs := make(map[string][]*APIBlock)
	abs["send"] = senders
	abs["receive"] = receivers
	return abs, nil
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

func messageSeri(message string) (types.Hash, []byte, error) {
	m, err := json.Marshal(message)
	if err != nil {
		return types.ZeroHash, nil, err
	}
	mHash, err := types.HashBytes(m)
	if err != nil {
		return types.ZeroHash, nil, err
	}
	return mHash, m, nil
}

func messageDeSeri(m []byte) (string, error) {
	var str string
	err := json.Unmarshal(m, &str)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (s *SMSApi) MessageHash(message string) (types.Hash, error) {
	mHash, _, err := messageSeri(message)
	if err != nil {
		return types.ZeroHash, err
	}
	return mHash, nil
}

func (s *SMSApi) MessageStore(message string) (types.Hash, error) {
	mHash, m, err := messageSeri(message)
	if err != nil {
		return types.ZeroHash, err
	}
	err = s.ledger.AddMessageInfo(mHash, m)
	if err != nil {
		return types.ZeroHash, err
	}
	return mHash, nil
}

func (s *SMSApi) MessageInfo(mHash types.Hash) (string, error) {
	m, err := s.ledger.GetMessageInfo(mHash)
	if err != nil {
		return "", err
	}
	str, err := messageDeSeri(m)
	if err != nil {
		return "", err
	}
	return str, nil
}
