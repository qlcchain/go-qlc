/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package services

import (
	"errors"
	"github.com/qlcchain/go-qlc/consensus/pov"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/log"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type PoVService struct {
	common.ServiceLifecycle
	povEngine *pov.PoVEngine
	logger    *zap.SugaredLogger
}

func NewPoVService(cfg *config.Config, accounts []*types.Account) *PoVService {
	povEngine, _ := pov.NewPovEngine(cfg, accounts)
	return &PoVService{
		povEngine: povEngine,
		logger:    log.NewLogger("pov_service"),
	}
}

func (pov *PoVService) GetPoVEngine() *pov.PoVEngine {
	return pov.povEngine
}

func (pov *PoVService) Init() error {
	if !pov.PreInit() {
		return errors.New("pre init fail")
	}
	defer pov.PostInit()
	return pov.povEngine.Init()
}

func (pov *PoVService) Start() error {
	if !pov.PreStart() {
		return errors.New("pre start fail")
	}
	defer pov.PostStart()

	return pov.povEngine.Start()
}

func (pov *PoVService) Stop() error {
	if !pov.PreStop() {
		return errors.New("pre stop fail")
	}
	defer pov.PostStop()

	return pov.povEngine.Stop()
}

func (pov *PoVService) Status() int32 {
	return pov.State()
}
