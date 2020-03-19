/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/topic"

	"github.com/qlcchain/go-qlc/ledger/process"

	"github.com/qlcchain/go-qlc/mock"

	"github.com/google/uuid"

	ctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
)

func setupTestCaseResend(t *testing.T) (func(t *testing.T), *ledger.Ledger, *ResendBlockService) {
	t.Parallel()
	dir := filepath.Join(config.QlcTestDataDir(), "rb", uuid.New().String())
	_ = os.RemoveAll(dir)
	cm := config.NewCfgManager(dir)
	_, _ = cm.Load()
	cc := ctx.NewChainContext(cm.ConfigFile)
	ls := NewLedgerService(cm.ConfigFile)

	err := ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	_ = ls.Start()
	_ = cc.Register(ctx.LedgerService, ls)
	rb := NewResendBlockService(cm.ConfigFile)

	return func(t *testing.T) {
		err := ls.Ledger.Close()
		if err != nil {
			t.Fatal(err)
		}
		_ = cc.Stop()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, ls.Ledger, rb
}

func TestResendBlockService(t *testing.T) {
	teardownTestCase, l, rb := setupTestCaseResend(t)
	defer teardownTestCase(t)
	err := rb.Init()
	if err != nil {
		t.Fatal(err)
	}

	lv := process.NewLedgerVerifier(l)
	bs, _ := mock.BlockChain(false)
	if err := lv.BlockCacheProcess(bs[0]); err != nil {
		t.Fatal(err)
	}
	t.Log("bs hash", bs[0].GetHash())
	for _, b := range bs[1:] {
		if err := lv.BlockCacheProcess(b); err != nil {
			t.Fatal(err)
		}
	}
	block := mock.StateBlockWithoutWork()
	block1 := mock.StateBlockWithoutWork()
	block2 := mock.StateBlockWithoutWork()
	rb.cc.EventBus().Publish(topic.EventAddBlockCache, block)
	rb.cc.EventBus().Publish(topic.EventAddBlockCache, block1)
	rb.cc.EventBus().Publish(topic.EventAddBlockCache, block2)
	if err := lv.BlockCacheProcess(block); err != nil {
		t.Fatal(err)
	}
	err = rb.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(5 * time.Second)
	err = rb.Stop()
	if err != nil {
		t.Fatal(err)
	}
	if rb.Status() != 6 {
		t.Fatal("stop failed.")
	}
	time.Sleep(100 * time.Millisecond)
}
