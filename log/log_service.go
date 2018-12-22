/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package log

import (
	"errors"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type LogService struct {
	common.ServiceLifecycle
	config *config.Config
}

func NewLogService(cfg *config.Config) *LogService {
	return &LogService{config: cfg}
}

func (ls *LogService) Init() error {
	if !ls.PreInit() {
		return errors.New("pre init fail")
	}
	defer ls.PostInit()

	return InitLog(ls.config)
}

func (ls *LogService) Start() error {
	if !ls.PreStart() {
		return errors.New("pre start fail")
	}
	defer ls.PostStart()

	return nil
}

func (ls *LogService) Stop() error {
	if !ls.PreStop() {
		return errors.New("pre stop fail")
	}
	defer ls.PostStop()

	return nil
}

func (ls *LogService) Status() int32 {
	return ls.State()
}
