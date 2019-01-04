package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/rpc"
)

var logger = log.NewLogger("main")

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	switch os.Args[1] {

	case "rpc":
		cfg := rpc.RPCConfig{
			RPCEnabled: true,
			WSEnabled:  true,
			IPCEnabled: true,
		}
		r := rpc.NewRPC(cfg)
		err := r.StartRPC()
		defer r.StopRPC()
		//defer r.stopInProcess()
		//defer r.stopHTTP()
		logger.Info("rpc started")
		if err != nil {
			logger.Error(err)
		}
		s := <-c

		fmt.Println("Got signal: ", s)
		return
	}
}
