/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package chain

import (
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/config"
)

type serviceManager interface {
	common.Service
	Register(name string, service common.Service) error
	UnRegister(name string) error
	AllServices() ([]common.Service, error)
	Service(name string) (common.Service, error)
	//Control
	ReloadService(name string) error
	RestartAll() error
	// config
	ConfigManager() (*config.CfgManager, error)
	Config() (*config.Config, error)
}
