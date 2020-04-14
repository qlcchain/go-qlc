/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestAutoReceiveService_Init(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	s := NewAutoReceiveService(cm.ConfigFile)
	err = s.Init()
	if err != nil {
		t.Fatal(err)
	}
	if s.State() != 2 {
		t.Fatal("auto receive init failed")
	}
	_ = s.Start()
	err = s.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if s.Status() != 6 {
		t.Fatal("stop failed.")
	}
}

func TestReceiveBlock(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := context.NewChainContext(cm.ConfigFile)

	l := ledger.NewLedger(cm.ConfigFile)
	verifier := process.NewLedgerVerifier(l)
	ctx := vmstore.NewVMContext(l, &contractaddress.MintageAddress)
	genesisInfos := config.GenesisInfos()

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

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(mock.MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}

	if err := cc.Init(func() error {
		logService := NewLogService(cm.ConfigFile)
		_ = cc.Register(context.LogService, logService)
		_ = logService.Init()
		ledgerService := NewLedgerService(cm.ConfigFile)
		_ = cc.Register(context.LedgerService, ledgerService)
		if rpcService, err := NewRPCService(cm.ConfigFile); err != nil {
			return err
		} else {
			_ = cc.Register(context.RPCService, rpcService)
		}

		autoReceiveService := NewAutoReceiveService(cm.ConfigFile)
		_ = cc.Register(context.AutoReceiveService, autoReceiveService)

		return nil
	}); err != nil {
		t.Fatal(err)
	}
	account1 := mock.Account1
	account2 := mock.Account2

	cc.SetAccounts([]*types.Account{account2})

	if err := cc.Start(); err != nil {
		t.Fatal(err)
	}
	cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	cc.EventBus().Publish(topic.EventAddP2PStream, &topic.EventAddP2PStreamMsg{PeerID: "123", PeerInfo: "234"})

	if sb, err := l.GenerateSendBlock(&types.StateBlock{
		Token:   config.ChainToken(),
		Address: account1.Address(),
		Link:    types.Hash(account2.Address()),
	}, types.Balance{Int: big.NewInt(1e8)}, account1.PrivateKey()); err != nil {
		t.Fatal(err)
	} else {
		if err := verifier.BlockProcess(sb); err != nil {
			t.Fatal(err)
		} else {
			//if err := ReceiveBlock(sb, account2, cc); err != nil {
			//	t.Fatal(err)
			//}
			_ = ReceiveBlock(sb, account2, cc)

			if s, err := cc.Service(context.AutoReceiveService); err != nil {
				t.Fatal(err)
			} else {
				as := s.(*AutoReceiveService)
				as.blockCache <- sb
				time.Sleep(time.Second)
				if err := as.Stop(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

}
