/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"testing"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/vmcontract/chaincontract"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func setPovStatus(l *ledger.Ledger, cc *qlcchainctx.ChainContext, t *testing.T) {
	block, td := mock.GeneratePovBlock(nil, 0)
	if err := l.AddPovBlock(block, td); err != nil {
		t.Fatal(err)
	}
	if err := l.AddPovBestHash(block.GetHeight(), block.GetHash()); err != nil {
		t.Fatal(err)
	}
	if err := l.SetPovLatestHeight(block.GetHeight()); err != nil {
		t.Fatal(err)
	}
	_ = cc.Init(func() error {
		chaincontract.InitChainContract()
		return nil
	})
	_ = cc.Start()
	cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	cc.EventBus().Publish(topic.EventAddP2PStream, &topic.EventAddP2PStreamMsg{PeerID: "123", PeerInfo: "234"})
}

func setLedgerStatus(l *ledger.Ledger, t *testing.T) {
	genesisInfos := config.GenesisInfos()

	ctx := vmstore.NewVMContext(l, &contractaddress.MintageAddress)
	for _, v := range genesisInfos {
		mb := v.Mintage
		gb := v.Genesis
		err := ctx.SetStorage(contractaddress.MintageAddress[:], gb.Token[:], gb.Data)
		if err != nil {
			t.Fatal(err)
		}
		verifier := process.NewLedgerVerifier(l)
		if b, err := l.HasStateBlock(mb.GetHash()); !b && err == nil {
			if err := l.AddStateBlock(&mb); err != nil {
				t.Fatal(err)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}

		if b, err := l.HasStateBlock(gb.GetHash()); !b && err == nil {
			if err := verifier.BlockProcess(&gb); err != nil {
				t.Fatal(err)
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := l.SaveStorage(vmstore.ToCache(ctx)); err != nil {
		t.Fatal(err)
	}
}
