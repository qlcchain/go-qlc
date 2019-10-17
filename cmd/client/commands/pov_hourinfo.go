package commands

import (
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
)

func addPovLastNHourInfoCmdByShell(parentCmd *ishell.Cmd) {
	beginTimeFlag := util.Flag{
		Name:  "beginTime",
		Must:  false,
		Usage: "begin time, unix seconds",
		Value: 0,
	}
	endTimeFlag := util.Flag{
		Name:  "endTime",
		Must:  false,
		Usage: "end time, unix seconds",
		Value: 0,
	}

	cmd := &ishell.Cmd{
		Name: "getLastNHourInfo",
		Help: "get last N hour info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{beginTimeFlag, endTimeFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			beginTime, _ := util.IntVar(c.Args, beginTimeFlag)
			endTime, _ := util.IntVar(c.Args, endTimeFlag)

			err := runPovLastNHourInfoCmd(beginTime, endTime)
			if err != nil {
				c.Println(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovLastNHourInfoCmd(beginTime int, endTime int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiGetLastNHourInfo)
	err = client.Call(rspInfo, "pov_getLastNHourInfo", beginTime, endTime)
	if err != nil {
		return err
	}

	fmt.Printf("AllBlockNum: %d, AllTxNum: %d\n",
		rspInfo.AllBlockNum, rspInfo.AllTxNum)

	fmt.Printf("Sha256dBlockNum: %d, X11BlockNum: %d, ScryptBlockNum: %d\n",
		rspInfo.Sha256dBlockNum, rspInfo.X11BlockNum, rspInfo.ScryptBlockNum)

	fmt.Printf("MaxBlockPerHour: %d, MinBlockPerHour: %d, AvgBlockPerHour: %d\n",
		rspInfo.MaxBlockPerHour, rspInfo.MinBlockPerHour, rspInfo.AvgBlockPerHour)

	fmt.Printf("MaxTxPerBlock: %d, MinTxPerBlock: %d, AvgTxPerBlock:%d\n",
		rspInfo.MaxTxPerBlock, rspInfo.MinTxPerBlock, rspInfo.AvgTxPerBlock)

	fmt.Printf("MaxTxPerHour: %d, MinTxPerHour: %d, AvgTxPerHour:%d\n",
		rspInfo.MaxTxPerHour, rspInfo.MinTxPerHour, rspInfo.AvgTxPerHour)

	fmt.Printf("%-4s %-6s %-10s %-13s %-13s %-7s %-7s %-7s %-7s\n",
		"Hour", "Blocks", "Txs", "RewardsM", "RewardsR", "SHA256D", "X11", "SCRYPT", "AUX")
	for _, hourItem := range rspInfo.HourItemList {
		fmt.Printf("%-4d %-6d %-10d %-13.2f %-13.2f %-7d %-7d %-7d %-7d\n",
			hourItem.Hour, hourItem.AllBlockNum, hourItem.AllTxNum,
			float64(hourItem.AllMinerReward.Uint64())/1e8, float64(hourItem.AllRepReward.Uint64())/1e8,
			hourItem.Sha256dBlockNum, hourItem.X11BlockNum, hourItem.ScryptBlockNum, hourItem.AuxBlockNum)
	}

	return nil
}
