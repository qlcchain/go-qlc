/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"github.com/pkg/errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

func NewWalletService(cfg *config.Config) *WalletService {
	return &WalletService{
		wallet: NewWalletStore(cfg),
	}
}

type WalletService struct {
	common.ServiceLifecycle
	wallet *WalletStore
}

func (ws *WalletService) Init() error {
	if !ws.PreInit() {
		return errors.New("pre init fail.")
	}
	defer ws.PostInit()

	return nil
}

func (ws *WalletService) Start() error {
	if !ws.PreStart() {
		return errors.New("pre start fail.")
	}
	defer ws.PostStart()

	return nil
}

func (ws *WalletService) Stop() error {
	if !ws.PreStop() {
		return errors.New("pre stop fail.")
	}
	defer ws.PostStop()

	return ws.wallet.Close()
}

func (ws *WalletService) Status() int32 {
	return ws.State()
}
