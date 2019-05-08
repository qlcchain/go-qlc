package api

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/relation"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"go.uber.org/zap"
)

type SMSApi struct {
	ledger    *ledger.Ledger
	vmContext *vmstore.VMContext
	relation  *relation.Relation
	logger    *zap.SugaredLogger
}

func NewSMSApi(ledger *ledger.Ledger, relation *relation.Relation) *SMSApi {
	return &SMSApi{ledger: ledger, vmContext: vmstore.NewVMContext(ledger),
		relation: relation, logger: log.NewLogger("api_sms")}
}

func phoneNumberSeri(number string) ([]byte, error) {
	if number == "" {
		return nil, nil
	}
	b := util.String2Bytes(number)
	return b, nil
}

func (s *SMSApi) getApiBlocksByHash(hashes []types.Hash) ([]*APIBlock, error) {
	ab := make([]*APIBlock, 0)
	for _, h := range hashes {
		block, err := s.ledger.GetStateBlock(h)
		if err != nil && err != ledger.ErrBlockNotFound {
			return nil, err
		}
		if block != nil {
			b, err := generateAPIBlock(s.vmContext, block)
			if err != nil {
				return nil, err
			}
			ab = append(ab, b)
		}
	}
	return ab, nil

}

func (s *SMSApi) PhoneBlocks(sender string) (map[string][]*APIBlock, error) {
	p, err := phoneNumberSeri(sender)
	if err != nil {
		return nil, errors.New("error phone number")
	}
	limit := 20
	offset := 0
	sHash, err := s.relation.PhoneBlocks(p, true, limit, offset)
	if err != nil {
		return nil, err
	}
	senders, err := s.getApiBlocksByHash(sHash)
	if err != nil {
		return nil, err
	}
	rHash, err := s.relation.PhoneBlocks(p, false, limit, offset)
	if err != nil {
		return nil, err
	}
	receivers, err := s.getApiBlocksByHash(rHash)
	if err != nil {
		return nil, err
	}
	abs := make(map[string][]*APIBlock)
	abs["send"] = senders
	abs["receive"] = receivers
	return abs, nil
}

func (s *SMSApi) MessageBlocks(hash types.Hash) ([]*APIBlock, error) {
	limit := 20
	offset := 0
	hashes, err := s.relation.MessageBlocks(hash, limit, offset)
	if err != nil {
		return nil, err
	}
	b, err := s.getApiBlocksByHash(hashes)
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
