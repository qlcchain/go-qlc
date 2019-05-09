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
	ConfigV2 `mapstructure:",squash"`
	DB       *DBConfig  `json:"db"`
	PoV      *PoVConfig `json:"pov"`
}

func DefaultConfigV3(dir string) (*ConfigV3, error) {
	var cfg ConfigV3
	cfg2, _ := DefaultConfigV2(dir)
	cfg.ConfigV2 = *cfg2
	cfg.Version = 3
	cfg.RPC.HttpVirtualHosts = []string{"*"}
	cfg.RPC.PublicModules = append(cfg.RPC.PublicModules, "pledge")

	cfg.DB = defaultDb(dir)
	cfg.PoV = defaultPoV()

	return &cfg, nil
}

var (
	relationDir = "relation"
	pwLen       = 16
)

type DBConfig struct {
	ConnectionString string `json:"connectionString"`
	Driver           string `json:"driver"`
}

type PoVConfig struct {
	BlockInterval int    `json:"blockInterval"`
	BlockSize     int    `json:"blockSize"`
	TargetCycle   int    `json:"targetCycle"`
	ForkHeight    int    `json:"forkHeight"`
	Coinbase      string `json:"coinbase"`
}

func defaultDb(dir string) *DBConfig {
	d := path.Join(dir, "ledger", relationDir, "index.db")
	pw := util.RandomFixedString(pwLen)

	//"postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	//postgres
	return &DBConfig{
		ConnectionString: fmt.Sprintf("file:%s?_auth&_auth_user=qlcchain&_auth_pass=%s", d, pw),
		Driver:           "sqlite3",
	}
}

func defaultPoV() *PoVConfig {
	return &PoVConfig{
		BlockInterval: 30,
		BlockSize:     4 * 1024 * 1024,
		TargetCycle:   20,
		ForkHeight:    3,
		Coinbase:      "",
	}
}
