package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/qlcchain/go-qlc/cmd/util"
	cutil "github.com/qlcchain/go-qlc/common/util"
)

func addPovDiffDayStatCmdByShell(parentCmd *ishell.Cmd) {
	dayFlag := util.Flag{
		Name:  "day",
		Must:  false,
		Usage: "index of pov day",
		Value: 0,
	}
	heightFlag := util.Flag{
		Name:  "height",
		Must:  false,
		Usage: "height of pov block",
		Value: 0,
	}

	cmd := &ishell.Cmd{
		Name: "getDiffDayStat",
		Help: "get difficulty day statistics info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{dayFlag, heightFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			day, _ := util.IntVar(c.Args, dayFlag)
			height, _ := util.IntVar(c.Args, heightFlag)

			err := runPovDiffDayStatCmd(day, height)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovDiffDayStatCmd(day int, height int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(types.PovDiffDayStat)
	if height > 0 {
		err = client.Call(rspInfo, "pov_getDiffDayStatByHeight", height)
	} else {
		err = client.Call(rspInfo, "pov_getDiffDayStat", day)
	}
	if err != nil {
		return err
	}

	fmt.Println(cutil.ToIndentString(rspInfo))

	return nil
}

func addPovMinerDayStatCmdByShell(parentCmd *ishell.Cmd) {
	dayFlag := util.Flag{
		Name:  "day",
		Must:  false,
		Usage: "index of pov day",
		Value: 0,
	}
	heightFlag := util.Flag{
		Name:  "height",
		Must:  false,
		Usage: "height of pov block",
		Value: 0,
	}

	cmd := &ishell.Cmd{
		Name: "getMinerDayStat",
		Help: "get miner day statistics info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{dayFlag, heightFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			day, _ := util.IntVar(c.Args, dayFlag)
			height, _ := util.IntVar(c.Args, heightFlag)

			err := runPovMinerDayStatCmd(day, height)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovMinerDayStatCmd(day int, height int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(types.PovMinerDayStat)
	if height > 0 {
		err = client.Call(rspInfo, "pov_getMinerDayStatByHeight", height)
	} else {
		err = client.Call(rspInfo, "pov_getMinerDayStat", day)
	}
	if err != nil {
		return err
	}

	//fmt.Println(cutil.ToIndentString(rspInfo))

	minerCount := 0
	repCount := 0
	for _, minerItem := range rspInfo.MinerStats {
		if minerItem.BlockNum > 0 {
			minerCount++
		}
		if minerItem.RepBlockNum > 0 {
			repCount++
		}
	}

	fmt.Printf("DayIndex:%d, TotalNum: %d, MinerNum: %d, RepNum: %d\n",
		rspInfo.DayIndex, len(rspInfo.MinerStats), minerCount, repCount)

	fmt.Printf("%-64s %-4s %-8s %-4s %-8s %-10s %-10s\n",
		"Address", "MBs", "MRs", "RBs", "RRs", "FirstH", "LastH")
	for minerAddr, minerItem := range rspInfo.MinerStats {
		fmt.Printf("%-64s %-4d %-8.2f %-4d %-8.2f %-10d %-10d\n",
			minerAddr,
			minerItem.BlockNum, float64(minerItem.RewardAmount.Uint64())/1e8,
			minerItem.RepBlockNum, float64(minerItem.RepReward.Uint64())/1e8,
			minerItem.FirstHeight, minerItem.LastHeight)
	}

	return nil
}
