package commands

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func repRewardHistory() {
	var accountP string

	if interactive {
		account := util.Flag{
			Name:  "account",
			Must:  true,
			Usage: "representative address hex string",
		}

		cmd := &ishell.Cmd{
			Name: "repRewardHistory",
			Help: "representative reward history info",
			Func: func(c *ishell.Context) {
				args := []util.Flag{account}
				if util.HelpText(c, args) {
					return
				}

				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				accountP = util.StringVar(c.Args, account)

				if err := repRewardHistoryAction(accountP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "repRewardHistory",
			Short: "representative reward history info",
			Run: func(cmd *cobra.Command, args []string) {
				err := repRewardHistoryAction(accountP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&accountP, "account", "", "representative address hex string")
		rootCmd.AddCommand(cmd)
	}
}

func repRewardHistoryAction(accountP string) error {
	if accountP == "" {
		return errors.New("invalid account value")
	}

	account, err := types.HexToAddress(accountP)
	if err != nil {
		return err
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspHistoryInfo := api.RepHistoryRewardInfo{}
	err = client.Call(&rspHistoryInfo, "rep_getHistoryRewardInfos", account)
	if err != nil {
		return err
	}

	gasAmount0 := new(big.Float).SetInt(rspHistoryInfo.AllRewardAmount.Int)
	gasAmount1 := new(big.Float).Quo(gasAmount0, new(big.Float).SetInt64(1e8))
	gasAmountF, _ := gasAmount1.Float64()

	fmt.Printf("AllRewardBlocks: %d\n", rspHistoryInfo.AllRewardBlocks)
	fmt.Printf("AllRewardAmount: %d (%f)\n", rspHistoryInfo.AllRewardAmount, gasAmountF)
	fmt.Printf("%-10s %-10s %-10s %-18s %-64s\n", "Start", "End", "Blocks", "Reward", "Beneficial")
	for _, info := range rspHistoryInfo.RewardInfos {
		fmt.Printf("%-10d %-10d %-10d %-18s %-64s\n", info.StartHeight, info.EndHeight, info.RewardBlocks, info.RewardAmount.String(), info.Beneficial)
	}

	return nil
}
