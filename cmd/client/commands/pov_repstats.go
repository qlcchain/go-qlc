package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
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

	rspInfo := make(map[types.Address]*api.PovRepStats, 0)
	err = client.Call(&rspInfo, "pov_getRepStats", repAddrStrList)
	if err != nil {
		return err
	}

	fmt.Printf("TotalCount: %d\n", len(rspInfo))

	fmt.Printf("%-64s %-10s %-13s\n", "Address", "Blocks", "Rewards")
	for repAddr, repItem := range rspInfo {
		fmt.Printf("%-64s %-10d %-13.2f\n",
			repAddr,
			repItem.MainBlockNum,
			float64(repItem.MainRewardAmount.Uint64())/1e8)
	}

	return nil
}
