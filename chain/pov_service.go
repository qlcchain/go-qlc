/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/consensus/pov"
	"github.com/qlcchain/go-qlc/log"
)

type PoVService struct {
	common.ServiceLifecycle
	povEngine *pov.PoVEngine
	logger    *zap.SugaredLogger
}

func NewPoVService(cfgFile string) *PoVService {
	povEngine, _ := pov.NewPovEngine(cfgFile, false)
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
