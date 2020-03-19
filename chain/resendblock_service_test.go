/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	ctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type mockLedgerService struct {
	common.ServiceLifecycle
}

func (m *mockLedgerService) Init() error {
	if !m.PreInit() {
		return errors.New("pre init fail")
	}
	defer m.PostInit()
	return nil
}

func (m *mockLedgerService) Start() error {
	if !m.PreStart() {
		return errors.New("pre start fail")
	}
	defer m.PostStart()
	return nil
}

func (m *mockLedgerService) Stop() error {
	if !m.PreStop() {
		return errors.New("pre stop fail")
	}
	defer m.PostStop()

	return nil
}

func (m *mockLedgerService) Status() int32 {
	return m.State()
}

func TestNewResendBlockService(t *testing.T) {
	t.Skip()
	dir := filepath.Join(config.QlcTestDataDir(), uuid.New().String())
	cm := config.NewCfgManager(dir)
	_, err := cm.Load()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	cc := ctx.NewChainContext(cm.ConfigFile)
	m := &mockLedgerService{}
	_ = m.Init()
	_ = m.Start()
	_ = cc.Register(ctx.LedgerService, m)

	ls := NewResendBlockService(cm.ConfigFile)

	err = ls.Init()
	if err != nil {
		t.Fatal(err)
	}
	if ls.State() != 2 {
		t.Fatal("resend init failed")
	}
	_ = ls.Start()
	err = ls.Stop()
	if err != nil {
		t.Fatal(err)
	}

	if ls.Status() != 6 {
		t.Fatal("stop failed.")
	}

	_ = m.Stop()
}
