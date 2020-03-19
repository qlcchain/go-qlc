package pov

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
)

func setupPovTxBlock2Ledger(l ledger.Store, povBlock *types.PovBlock) {
	for _, txPov := range povBlock.Body.Txs {
		if txPov.Block != nil {
			_ = l.AddStateBlock(txPov.Block)
		}
	}
}
