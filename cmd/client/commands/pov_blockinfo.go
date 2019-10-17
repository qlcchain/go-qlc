package commands

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
	"strconv"
)

func addPovHeaderInfoCmdByShell(parentCmd *ishell.Cmd) {
	idFlag := util.Flag{
		Name:  "id",
		Must:  true,
		Usage: "height or hash of pov block",
		Value: "-1",
	}

	cmd := &ishell.Cmd{
		Name: "getHeader",
		Help: "get header info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{idFlag}
			if util.HelpText(c, args) {
				return
			}
			idStr := util.StringVar(c.Args, idFlag)

			err := runPovHeaderInfoCmd(idStr)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovHeaderInfoCmd(idStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiHeader)

	if len(idStr) == 64 {
		err = client.Call(rspInfo, "pov_getHeaderByHash", idStr)
		if err != nil {
			return err
		}
	} else {
		var height int64
		if idStr == "latest" {
			height = int64(-1)
		} else {
			height, err = strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return err
			}
		}

		if height < 0 {
			err = client.Call(rspInfo, "pov_getLatestHeader")
		} else {
			err = client.Call(rspInfo, "pov_getHeaderByHeight", height)
		}
		if err != nil {
			return err
		}
	}

	fmt.Printf("Hash: %s\n", rspInfo.BasHdr.Hash)
	fmt.Printf("Height: %d\n", rspInfo.BasHdr.Height)
	fmt.Printf("Difficulty: %s\n", formatPovDifficulty(rspInfo.NormDifficulty))
	if rspInfo.CbTx != nil {
		fmt.Printf("TxNum: %d\n", rspInfo.CbTx.TxNum)
		fmt.Printf("StateHash: %s\n", rspInfo.CbTx.StateHash)
	}

	fmt.Printf("AlgoName: %s\n", rspInfo.AlgoName)
	fmt.Printf("AlgoEfficiency: %d\n", rspInfo.AlgoEfficiency)
	fmt.Printf("AlgoDifficulty: %s\n", formatPovDifficulty(rspInfo.AlgoDifficulty))

	if rspInfo.AuxHdr == nil {
		fmt.Printf("AuxPoW: false\n")
	} else {
		fmt.Printf("AuxPoW: true\n")
		fmt.Printf("ParentHash: %s\n", rspInfo.AuxHdr.ParentHash)
	}

	fmt.Printf("Miner: %s\n", rspInfo.GetMinerAddr())
	fmt.Printf("Miner Reward: %s QGAS\n", formatPovReward(rspInfo.GetMinerReward()))
	fmt.Printf("Representative Reward: %s QGAS\n", formatPovReward(rspInfo.GetRepReward()))

	return nil
}
