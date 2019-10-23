/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/consensus/dpos"
)

type ConsensusService struct {
	common.ServiceLifecycle
	cfgFile string
	c       *consensus.Consensus
}

func NewConsensusService(cfgFile string) *ConsensusService {
	return &ConsensusService{
		cfgFile: cfgFile,
	}
}

func (cs *ConsensusService) Init() error {
	dPoS := dpos.NewDPoS(cs.cfgFile)

	cs.PreInit()
	cs.c = consensus.NewConsensus(dPoS, cs.cfgFile)
	cs.c.Init()
	cs.PostInit()

	return nil
}

func (cs *ConsensusService) Start() error {
	cs.PreStart()
	cs.c.Start()
	cs.PostStart()

	return nil
}

func (cs *ConsensusService) Stop() error {
	cs.PreStop()
	cs.c.Stop()
	cs.PostStop()

	return nil
}

func (cs *ConsensusService) Status() int32 {
	return cs.State()
}
