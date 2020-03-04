/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"

	"github.com/qlcchain/go-qlc/p2p"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/log"
)

type P2PService struct {
	common.ServiceLifecycle
	p2p    *p2p.QlcService
	logger *zap.SugaredLogger
}

func NewP2PService(cfgFile string) (*P2PService, error) {
	p, err := p2p.NewQlcService(cfgFile)
	if err != nil {
		return nil, err
	}
	return &P2PService{p2p: p, logger: log.NewLogger("p2p_service")}, nil
}

func (p *P2PService) Init() error {
	if !p.PreInit() {
		return errors.New("pre init fail")
	}
	defer p.PostInit()

	return nil
}

func (p *P2PService) Start() error {
	if !p.PreStart() {
		return errors.New("pre start fail")
	}
	err := p.p2p.Start()
	if err != nil {
		p.logger.Error(err)
		return err
	}
	p.PostStart()
	return nil
}

func (p *P2PService) Stop() error {
	if !p.PreStop() {
		return errors.New("pre stop fail")
	}
	defer p.PostStop()

	p.p2p.Stop()
	p.logger.Info("p2p stopped")
	return nil
}

func (p *P2PService) Status() int32 {
	return p.State()
}

func (p *P2PService) Node() *p2p.QlcService {
	return p.p2p
}
