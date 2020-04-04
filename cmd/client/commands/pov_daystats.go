package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
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
	filterFlag := util.Flag{
		Name:  "filter",
		Must:  false,
		Usage: "filter type, etc all/miner/rep",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getMinerDayStat",
		Help: "get miner day statistics info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{dayFlag, heightFlag, filterFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			day, _ := util.IntVar(c.Args, dayFlag)
			height, _ := util.IntVar(c.Args, heightFlag)
			filter := util.StringVar(c.Args, filterFlag)

			err := runPovMinerDayStatCmd(day, height, filter)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovMinerDayStatCmd(day int, height int, filter string) error {
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
	minerBlkCount := uint32(0)
	minerRwdAmount := uint64(0)
	repCount := 0
	repRwdAmount := uint64(0)
	for _, minerItem := range rspInfo.MinerStats {
		if minerItem.BlockNum > 0 {
			minerCount++
			minerBlkCount += minerItem.BlockNum
			minerRwdAmount += minerItem.RewardAmount.Uint64()
		}
		if minerItem.RepBlockNum > 0 {
			repCount++
			repRwdAmount += minerItem.RepReward.Uint64()
		}
	}

	fmt.Printf("DayIndex: %d, TotalNum: %d\n", rspInfo.DayIndex, len(rspInfo.MinerStats))

	fmt.Printf("MinerNum: %d, MinerBlockNum: %d, MinerRewardAmount: %.2f\n",
		minerCount, minerBlkCount, float64(minerRwdAmount)/1e8)

	fmt.Printf("RepNum: %d, RepRewardAmount: %.2f\n",
		repCount, float64(repRwdAmount)/1e8)

	fmt.Printf("%-64s %-4s %-8s %-4s %-8s %-10s %-10s\n",
		"Address", "MBs", "MRs", "RBs", "RRs", "FirstH", "LastH")
	for minerAddr, minerItem := range rspInfo.MinerStats {
		if filter != "" {
			if filter == "miner" && minerItem.BlockNum == 0 {
				continue
			}
			if filter == "rep" && minerItem.RepBlockNum == 0 {
				continue
			}
		}

		fmt.Printf("%-64s %-4d %-8.2f %-4d %-8.2f %-10d %-10d\n",
			minerAddr,
			minerItem.BlockNum, float64(minerItem.RewardAmount.Uint64())/1e8,
			minerItem.RepBlockNum, float64(minerItem.RepReward.Uint64())/1e8,
			minerItem.FirstHeight, minerItem.LastHeight)
	}

	return nil
}
