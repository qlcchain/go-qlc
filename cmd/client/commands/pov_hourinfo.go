package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
)

func addPovLastNHourInfoCmdByShell(parentCmd *ishell.Cmd) {
	endHeightFlag := util.Flag{
		Name:  "endHeight",
		Must:  false,
		Usage: "end height, 0 is latest",
		Value: 0,
	}
	hourSpanFlag := util.Flag{
		Name:  "hourSpan",
		Must:  false,
		Usage: "hour span, [2 ~ 24]",
		Value: 0,
	}

	cmd := &ishell.Cmd{
		Name: "getLastNHourInfo",
		Help: "get last N hour info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{endHeightFlag, hourSpanFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			endHeight, _ := util.IntVar(c.Args, endHeightFlag)
			hourSpan, _ := util.IntVar(c.Args, hourSpanFlag)

			err := runPovLastNHourInfoCmd(endHeight, hourSpan)
			if err != nil {
				c.Println(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovLastNHourInfoCmd(endHeight int, hourSpan int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiGetLastNHourInfo)
	err = client.Call(rspInfo, "pov_getLastNHourInfo", endHeight, hourSpan)
	if err != nil {
		return err
	}

	fmt.Printf("AllBlockNum: %d, AllTxNum: %d\n",
		rspInfo.AllBlockNum, rspInfo.AllTxNum)

	fmt.Printf("Sha256dBlockNum: %d, X11BlockNum: %d, ScryptBlockNum: %d\n",
		rspInfo.Sha256dBlockNum, rspInfo.X11BlockNum, rspInfo.ScryptBlockNum)

	fmt.Printf("AuxBlockNum: %d\n", rspInfo.AuxBlockNum)

	fmt.Printf("\n")
	fmt.Printf("MaxBlockPerHour: %d, MinBlockPerHour: %d, AvgBlockPerHour: %d\n",
		rspInfo.MaxBlockPerHour, rspInfo.MinBlockPerHour, rspInfo.AvgBlockPerHour)

	fmt.Printf("MaxTxPerBlock: %d, MinTxPerBlock: %d, AvgTxPerBlock: %d\n",
		rspInfo.MaxTxPerBlock, rspInfo.MinTxPerBlock, rspInfo.AvgTxPerBlock)

	fmt.Printf("MaxTxPerHour: %d, MinTxPerHour: %d, AvgTxPerHour: %d\n",
		rspInfo.MaxTxPerHour, rspInfo.MinTxPerHour, rspInfo.AvgTxPerHour)

	fmt.Printf("\n")
	fmt.Printf("%-4s %-6s %-8s %-13s %-13s %-7s %-7s %-7s %-7s\n",
		"Hour", "Blocks", "Txs", "RewardsM", "RewardsR", "SHA256D", "X11", "SCRYPT", "AUX")
	for _, hourItem := range rspInfo.HourItemList {
		fmt.Printf("%-4d %-6d %-8d %-13.2f %-13.2f %-7d %-7d %-7d %-7d\n",
			hourItem.Hour, hourItem.AllBlockNum, hourItem.AllTxNum,
			float64(hourItem.AllMinerReward.Uint64())/1e8, float64(hourItem.AllRepReward.Uint64())/1e8,
			hourItem.Sha256dBlockNum, hourItem.X11BlockNum, hourItem.ScryptBlockNum, hourItem.AuxBlockNum)
	}

	return nil
}
