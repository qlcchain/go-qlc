// +build !testnet

package contract

import (
	"errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PrivacyDemoKVSet struct {
	BaseContract
}

func (s *PrivacyDemoKVSet) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	return nil, nil, errors.New("contract not supported")
}
