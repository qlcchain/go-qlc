/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/consensus/dpos"

	"github.com/qlcchain/go-qlc/common/event"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
)

type ConsensusService struct {
	common.ServiceLifecycle
	cfg      *config.Config
	accounts []*types.Account
	eb       event.EventBus
	c        *consensus.Consensus
}

func NewConsensusService(cfg *config.Config, accounts []*types.Account) *ConsensusService {
	return &ConsensusService{
		cfg:      cfg,
		accounts: accounts,
		eb:       event.GetEventBus(cfg.LedgerDir()),
	}
}

func (cs *ConsensusService) Init() error {
	dPoS := dpos.NewDPoS(cs.cfg, cs.accounts, cs.eb)

	cs.PreInit()
	cs.c = consensus.NewConsensus(dPoS, cs.cfg, cs.eb)
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
