// +build !testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package config

import (
	"fmt"
	"path"

	"github.com/qlcchain/go-qlc/common/util"
)

type ConfigV3 struct {
	ConfigV2
	DB *DBConfig `json:"db"`
}

type DBConfig struct {
	ConnectionString string `json:"connectionString"`
	Driver           string `json:"driver"`
}

func DefaultConfigV3(dir string) (*ConfigV3, error) {
	var cfg ConfigV3
	cfgv2, _ := DefaultConfigV2(dir)
	cfg.ConfigV2 = *cfgv2
	cfg.RPC.HttpVirtualHosts = []string{"*"}
	pw := util.RandomFixedString(16)

	//"postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	//postgres
	d := path.Join(dir, "ledger", "relation", "index.db")
	cfg.DB = &DBConfig{
		ConnectionString: fmt.Sprintf("file:%s?_auth&_auth_user=qlcchain&_auth_pass=%s", d, pw),
		Driver:           "sqlite3",
	}

	return &cfg, nil
}
