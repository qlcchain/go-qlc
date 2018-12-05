/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"flag"
	"fmt"

	"github.com/qlcchain/go-qlc/config"
)

func getConfig() (*config.Config, error) {
	defaultConfig, err := config.InitConfig()
	if err != nil {
		return nil, err
	}
	//flag.UintVar(&defaultConfig.RPC.Port, "RPC_PORT", 29999, "the port of rpc")
	flag.StringVar(&defaultConfig.ID.PeerID, "PeerID", defaultConfig.ID.PeerID, "network id")
	flag.Parse()
	return defaultConfig, nil
}

func main() {
	cfg, err := getConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cfg.ID.PeerID)
}
