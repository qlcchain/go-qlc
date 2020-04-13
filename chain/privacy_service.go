/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"errors"

	"github.com/qlcchain/go-qlc/chain/context"

	"go.uber.org/zap"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/privacy"
)

type PrivacyService struct {
	common.ServiceLifecycle
	controller *privacy.Controller
	logger     *zap.SugaredLogger
}

func NewPrivacyService(cfgFile string) *PrivacyService {
	cc := context.NewChainContext(cfgFile)

	ctrl := privacy.NewController(cc)
	return &PrivacyService{
		controller: ctrl,
		logger:     log.NewLogger("privacy_service"),
	}
}

func (s *PrivacyService) Init() error {
	if !s.PreInit() {
		return errors.New("pre init fail")
	}
	defer s.PostInit()
	return s.controller.Init()
}

func (s *PrivacyService) Start() error {
	if !s.PreStart() {
		return errors.New("pre start fail")
	}
	defer s.PostStart()

	return s.controller.Start()
}

func (s *PrivacyService) Stop() error {
	if !s.PreStop() {
		return errors.New("pre stop fail")
	}
	defer s.PostStop()

	return s.controller.Stop()
}

func (s *PrivacyService) Status() int32 {
	return s.State()
}
