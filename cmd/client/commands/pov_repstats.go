package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPovRepInfoCmdByShell(parentCmd *ishell.Cmd) {
	repFlag := util.Flag{
		Name:  "reps",
		Must:  false,
		Usage: "addresses of representatives",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getRepStats",
		Help: "get representative statistic info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{repFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			repAddrStrList := util.StringSliceVar(c.Args, repFlag)

			err := runPovRepInfoCmd(repAddrStrList)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovRepInfoCmd(repAddrStrList []string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovRepStats)
	err = client.Call(&rspInfo, "pov_getRepStats", repAddrStrList)
	if err != nil {
		return err
	}

	fmt.Printf("TotalBlockNum: %d, LatestBlockHeight: %d\n", rspInfo.TotalBlockNum, rspInfo.LatestBlockHeight)
	fmt.Printf("TotalReward: %s, TotalPeriod: %d\n", formatPovReward(rspInfo.TotalRewardAmount), rspInfo.TotalPeriod)

	fmt.Printf("%-64s %-10s %-10s %-13s %-10s %-10s %-13s %-10s\n", "Address", "MBlocks", "MPeriod", "MRewards", "SBlocks", "SPeriod", "SRewards", "IsOnline")
	for repAddr, repItem := range rspInfo.RepStats {
		fmt.Printf("%-64s %-10d %-10d %-13.2f %-10d %-10d %-13.2f %-10t\n",
			repAddr,
			repItem.MainBlockNum,
			repItem.MainOnlinePeriod,
			float64(repItem.MainRewardAmount.Uint64())/1e8,
			repItem.StableBlockNum,
			repItem.StableOnlinePeriod,
			float64(repItem.StableRewardAmount.Uint64())/1e8,
			repItem.IsOnline)
	}

	return nil
}
