/*
 * Copyright (c) 2020 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	qlcchainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/ledger/process"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract"
)

func setupMintageAPI(t *testing.T) (func(t *testing.T), *process.LedgerVerifier, *MintageAPI) {
	dir := filepath.Join(config.QlcTestDataDir(), "api", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := qlcchainctx.NewChainContext(cm.ConfigFile)
	l := ledger.NewLedger(cm.ConfigFile)

	verifier := process.NewLedgerVerifier(l)
	setPovStatus(l, cc, t)
	setLedgerStatus(l, t)

	api := NewMintageApi(cc.ConfigFile(), l)

	var blocks []*types.StateBlock
	if err := json.Unmarshal([]byte(MockBlocks), &blocks); err != nil {
		t.Fatal(err)
	}

	for i := range blocks {
		block := blocks[i]
		if err := verifier.BlockProcess(block); err != nil {
			t.Fatal(err)
		}
	}

	return func(t *testing.T) {
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
	}, verifier, api
}

func TestNewMintageApi(t *testing.T) {
	testcase, verifier, api := setupMintageAPI(t)
	defer testcase(t)

	a := account1.Address()

	tm, err := api.l.GetTokenMeta(a, config.ChainToken())
	if err != nil {
		t.Fatal(err)
	}

	param := &MintageParams{
		SelfAddr:    a,
		PrevHash:    tm.Header,
		TokenName:   "Test",
		TokenSymbol: "QT",
		TotalSupply: "100",
		Decimals:    8,
		Beneficial:  a,
		NEP5TxId:    mock.Hash().String(),
	}

	if data, err := api.GetMintageData(param); err != nil {
		t.Fatal(err)
	} else if len(data) == 0 {
		t.Fatal("invalid mintage data")
	}
	contract.SetMinMintageTime(0, 0, 0, 0, 0, 1)

	if blk, err := api.GetMintageBlock(param); err != nil {
		t.Fatal(err)
	} else {
		txHash := blk.GetHash()
		blk.Signature = account1.Sign(txHash)
		if err := verifier.BlockProcess(blk); err != nil {
			t.Fatal(err)
		}

		if rxBlk, err := api.GetRewardBlock(blk); err != nil {
			t.Fatal(err)
		} else {
			txHash := rxBlk.GetHash()
			rxBlk.Signature = account1.Sign(txHash)
			if err := verifier.BlockProcess(rxBlk); err != nil {
				t.Fatal(err)
			}

			ti, err := api.ParseTokenInfo(rxBlk.Data)
			if err != nil {
				t.Fatal(err)
			}

			if data, err := api.GetWithdrawMintageData(ti.TokenId); err != nil {
				t.Fatal(err)
			} else if len(data) == 0 {
				t.Fatal("invalid GetWithdrawMintageData")
			}

			time.Sleep(2 * time.Second)

			if blk, err := api.GetWithdrawMintageBlock(&WithdrawParams{
				SelfAddr: a,
				TokenId:  ti.TokenId,
			}); err != nil {
				t.Fatal(err)
			} else {
				txHash := blk.GetHash()
				blk.Signature = account1.Sign(txHash)
				if err := verifier.BlockProcess(blk); err != nil {
					t.Fatal(err)
				}

				if rxBlk, err := api.GetWithdrawRewardBlock(blk); err != nil {
					t.Fatal(err)
				} else {
					txHash := rxBlk.GetHash()
					rxBlk.Signature = account1.Sign(txHash)
					if err := verifier.BlockProcess(rxBlk); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}
