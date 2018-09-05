package main

import "github.com/qlcchain/go-qlc/common"

func main() {
	logger := common.NewLogger("module1")
	logger.Info("ttt1")

	logger2 := common.NewLogger("module2")

	logger2.Info("ttt2")
}
