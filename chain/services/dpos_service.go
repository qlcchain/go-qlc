/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package services

import (
	"errors"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/p2p"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type DPosService struct {
	common.ServiceLifecycle
	dpos   *consensus.DPoS
	logger *zap.SugaredLogger
}

func NewDPosService(cfg *config.Config, p2p p2p.Service, account types.Address, password string) *DPosService {
	dPoS, _ := consensus.NewDPoS(cfg, p2p, account, password)
	return &DPosService{
		dpos:   dPoS,
		logger: log.NewLogger("ledger_service"),
	}
}

func (dps *DPosService) DPos() *consensus.DPoS {
	return dps.dpos
	return dps.dpos
}

func (dps *DPosService) Init() error {
	if !dps.PreInit() {
		return errors.New("pre init fail")
	}
	defer dps.PostInit()
	return dps.dpos.Init()
}

func (dps *DPosService) Start() error {
	if !dps.PreStart() {
		return errors.New("pre start fail")
	}
	defer dps.PostStart()

	return dps.dpos.Start()
}

func (dps *DPosService) Stop() error {
	if !dps.PreStop() {
		return errors.New("pre stop fail")
	}
	defer dps.PostStop()

	return dps.dpos.Stop()
}

func (dps *DPosService) Status() int32 {
	return dps.State()
}
