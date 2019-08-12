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

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/miner"
	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type MinerService struct {
	common.ServiceLifecycle
	miner  *miner.Miner
	logger *zap.SugaredLogger
}

func NewMinerService(cfg *config.Config, povEngine *pov.PoVEngine) *MinerService {
	miner := miner.NewMiner(cfg, povEngine)
	return &MinerService{
		miner:  miner,
		logger: log.NewLogger("miner_service"),
	}
}

func (ms *MinerService) GetMiner() *miner.Miner {
	return ms.miner
}

func (ms *MinerService) Init() error {
	if !ms.PreInit() {
		return errors.New("pre init fail")
	}
	defer ms.PostInit()
	return ms.miner.Init()
}

func (ms *MinerService) Start() error {
	if !ms.PreStart() {
		return errors.New("pre start fail")
	}
	defer ms.PostStart()

	return ms.miner.Start()
}

func (ms *MinerService) Stop() error {
	if !ms.PreStop() {
		return errors.New("pre stop fail")
	}
	defer ms.PostStop()

	return ms.miner.Stop()
}

func (ms *MinerService) Status() int32 {
	return ms.State()
}
