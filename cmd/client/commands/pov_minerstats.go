package commands

import (
	"fmt"
	"sort"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPovMinerInfoCmdByShell(parentCmd *ishell.Cmd) {
	minerFlag := util.Flag{
		Name:  "miners",
		Must:  false,
		Usage: "addresses of miners",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getMinerStats",
		Help: "get miner statistic info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{minerFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			minerAddrStrList := util.StringSliceVar(c.Args, minerFlag)

			err := runPovMinerInfoCmd(minerAddrStrList)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovMinerInfoCmd(minerAddrStrList []string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovMinerStats)
	err = client.Call(rspInfo, "pov_getMinerStats", minerAddrStrList)
	if err != nil {
		return err
	}

	var sortMiners []*api.PovMinerStatItem
	for _, miner := range rspInfo.MinerStats {
		sortMiners = append(sortMiners, miner)
	}
	sort.Slice(sortMiners, func(i, j int) bool {
		if sortMiners[i].LastBlockHeight > sortMiners[j].LastBlockHeight {
			return true
		}
		return false
	})

	fmt.Printf("TotalBlockNum: %d, LatestBlockHeight: %d\n", rspInfo.TotalBlockNum, rspInfo.LatestBlockHeight)
	fmt.Printf("TotalRewardAmount: %s, TotalMinerReward: %s, TotalRepReward: %s\n",
		formatPovReward(rspInfo.TotalRewardAmount),
		formatPovReward(rspInfo.TotalMinerReward),
		formatPovReward(rspInfo.TotalRepReward))
	fmt.Printf("TotalMinerCount: %d, LastDayOnlineCount: %d, LastHourOnlineCount: %d\n", rspInfo.MinerCount, rspInfo.DayOnlineCount, rspInfo.HourOnlineCount)

	fmt.Printf("%-64s %-6s %-10s %-13s %-10s %-10s\n", "Address", "Online", "MBlocks", "MRewards", "FirstH", "LastH")
	for _, minerItem := range sortMiners {
		isDayInt := 0
		isHourInt := 0
		if minerItem.IsDayOnline {
			isDayInt = 1
		}
		if minerItem.IsHourOnline {
			isHourInt = 1
		}
		fmt.Printf("%-64s %-6s %-10d %-13.2f %-10d %-10d\n",
			minerItem.Account,
			fmt.Sprintf("%d/%d", isDayInt, isHourInt),
			minerItem.MainBlockNum,
			float64(minerItem.MainRewardAmount.Uint64())/1e8,
			minerItem.FirstBlockHeight, minerItem.LastBlockHeight)
	}

	return nil
}
