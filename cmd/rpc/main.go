package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/consensus"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc"
)

var logger = log.NewLogger("main")

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	switch os.Args[1] {

	case "rpc":
		cm := config.NewCfgManager(config.DefaultConfigFile())
		cfg, err := cm.Load()
		if cfg.RPC.Enable == false {
			return
		}

		dp := &consensus.DposService{}
		rs := rpc.NewRPCService(cfg, dp)
		err = rs.Init()
		if err != nil {
			logger.Fatal(err)
		}
		err = rs.Start()
		if err != nil {
			logger.Fatal(err)
		}

		defer rs.Stop()
		logger.Info("rpc started")
		s := <-c
		fmt.Println("Got signal: ", s)
	}
}
