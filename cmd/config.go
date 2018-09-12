/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package main

import (
	"github.com/qlcchain/go-qlc/config"
	"fmt"
	"flag"
)

func getConfig() *config.Config {
	var defaultConfig = config.DefaultlConfig
	//flag.UintVar(&defaultConfig.RPC.Port, "RPC_PORT", 29999, "the port of rpc")
	flag.UintVar(&defaultConfig.RPC.Port, "RPC_PORT", defaultConfig.RPC.Port, "the port of rpc")
	flag.Parse()
	return defaultConfig
}

func main() {
	cfg := getConfig()
	fmt.Println(cfg.RPC.Port)
}
