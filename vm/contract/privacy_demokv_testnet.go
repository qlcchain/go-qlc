// +build testnet

package contract

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PrivacyDemoKVSet struct {
	BaseContract
}

func (s *PrivacyDemoKVSet) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	log := log.NewLogger("PrivacyDemoKVSet")

	log.Debugf("ProcessSend block %+v", block)

	para := new(abi.PrivacyDemoKVABISetPara)
	err := abi.PrivacyDemoKVABI.UnpackMethod(para, abi.MethodNamePrivacyDemoKVSet, block.GetPayload())
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	var key []byte
	key = append(key, abi.PrivacyDemoKVStorageTypeKV)
	key = append(key, para.Key...)
	err = ctx.SetStorage(types.PrivacyDemoKVAddress[:], key, para.Value)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, nil
}
