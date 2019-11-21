package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addMinerRewardInfoCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "account address hex string",
	}

	cmd := &ishell.Cmd{
		Name: "getRewardInfo",
		Help: "miner get reward info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{address}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			addressP := util.StringVar(c.Args, address)

			if err := minerRewardInfoAction(addressP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func minerRewardInfoAction(addressP string) error {
	if addressP == "" {
		return errors.New("invalid addressP value")
	}
	minerAddr, _ := types.HexToAddress(addressP)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	minerRsp := new(api.PovMinerStats)
	err = client.Call(minerRsp, "pov_getMinerStats", []types.Address{minerAddr})
	if err != nil {
		return err
	}
	minerInfo := minerRsp.MinerStats[minerAddr]
	if minerInfo == nil {
		return errors.New("miner not exist")
	}
	fmt.Printf("Miner Statistic Info:\n")
	fmt.Printf("MainBlockNum:       %d\n", minerInfo.MainBlockNum)
	fmt.Printf("MainRewardAmount:   %s\n", txFormatBalance(minerInfo.MainRewardAmount))
	fmt.Printf("StableBlockNum:     %d\n", minerInfo.StableBlockNum)
	fmt.Printf("StableRewardAmount: %s\n", txFormatBalance(minerInfo.StableRewardAmount))

	// history reward info
	histInfo := new(api.MinerHistoryRewardInfo)
	err = client.Call(histInfo, "miner_getRewardHistory", minerAddr)
	fmt.Printf("\nReward History Info:\n")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("LastRewardTime: %s\n", time.Unix(histInfo.LastRewardTime, 0))
		fmt.Printf("LastEndHeight:  %d\n", histInfo.LastEndHeight)
		fmt.Printf("RewardBlocks:   %d\n", histInfo.RewardBlocks)
		fmt.Printf("RewardAmount:   %s\n", txFormatBalance(histInfo.RewardAmount))
	}

	// avail reward info
	availInfo := new(api.MinerAvailRewardInfo)
	err = client.Call(availInfo, "miner_getAvailRewardInfo", minerAddr)
	fmt.Printf("\nReward Available Info:\n")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("AvailStartHeight:  %d\n", availInfo.AvailStartHeight)
		fmt.Printf("AvailEndHeight:    %d\n", availInfo.AvailEndHeight)
		fmt.Printf("AvailRewardBlocks: %d\n", availInfo.AvailRewardBlocks)
		fmt.Printf("AvailRewardAmount: %s\n", txFormatBalance(availInfo.AvailRewardAmount))
		fmt.Printf("NodeRewardHeight:  %d\n", availInfo.NodeRewardHeight)
		fmt.Printf("NeedCallReward:    %t\n", availInfo.NeedCallReward)
	}

	return nil
}
