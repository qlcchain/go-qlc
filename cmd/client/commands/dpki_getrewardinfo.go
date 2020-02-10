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

func addDpkiRewardInfoCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "account address hex string",
	}

	cmd := &ishell.Cmd{
		Name: "getRewardInfo",
		Help: "get verifier reward info",
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

			if err := dpkiRewardInfoAction(addressP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func dpkiRewardInfoAction(addressP string) error {
	if addressP == "" {
		return errors.New("invalid addressP value")
	}
	rwdAddr, _ := types.HexToAddress(addressP)

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	// history reward info
	histInfo := new(api.PKDHistoryRewardInfo)
	err = client.Call(histInfo, "dpki_getRewardHistory", rwdAddr)
	fmt.Printf("\nReward History Info:\n")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("LastEndHeight:     %d\n", histInfo.LastEndHeight)
		fmt.Printf("LastBeneficial:    %s\n", histInfo.LastBeneficial)
		fmt.Printf("LastRewardTime:    %s\n", time.Unix(histInfo.LastRewardTime, 0))
		fmt.Printf("TotalRewardAmount: %s\n", txFormatBalance(histInfo.RewardAmount))
	}

	// avail reward info
	availInfo := new(api.PKDAvailRewardInfo)
	err = client.Call(availInfo, "dpki_getAvailRewardInfo", rwdAddr)
	fmt.Printf("\nReward Available Info:\n")
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("LastEndHeight:     %d\n", availInfo.LastEndHeight)
		fmt.Printf("AvailEndHeight:    %d\n", availInfo.AvailEndHeight)
		fmt.Printf("AvailRewardAmount: %s\n", txFormatBalance(availInfo.AvailRewardAmount))
		fmt.Printf("LatestBlockHeight: %d\n", availInfo.LatestBlockHeight)
		fmt.Printf("NodeRewardHeight:  %d\n", availInfo.NodeRewardHeight)
		fmt.Printf("NeedCallReward:    %t\n", availInfo.NeedCallReward)
	}

	return nil
}
