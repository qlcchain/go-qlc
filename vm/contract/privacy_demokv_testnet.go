// +build testnet

package contract

import (
	"encoding/hex"
	"errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PrivacyDemoKVSet struct {
	BaseContract
}

func (s *PrivacyDemoKVSet) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	logger := log.NewLogger("PrivacyDemoKVSet")

	if len(block.GetPayload()) == 0 {
		return nil, nil, errors.New("payload is nil")
	}

	para := new(abi.PrivacyDemoKVABISetPara)
	err := abi.PrivacyDemoKVABI.UnpackMethod(para, abi.MethodNamePrivacyDemoKVSet, block.GetPayload())
	if err != nil {
		logger.Errorf("failed to unpack payload:%s", hex.EncodeToString(block.GetPayload()))
		return nil, nil, ErrUnpackMethod
	}

	if len(para.Key) == 0 {
		return nil, nil, errors.New("invalid Key para")
	}
	if len(para.Value) == 0 {
		return nil, nil, errors.New("invalid Value para")
	}

	var key []byte
	key = append(key, abi.PrivacyDemoKVStorageTypeKV)
	key = append(key, para.Key...)
	err = ctx.SetStorage(contractaddress.PrivacyDemoKVAddress[:], key, para.Value)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}
