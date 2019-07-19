package commands

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
	"math/big"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func minerHistory() {
	var coinbaseP string

	if interactive {
		coinbase := util.Flag{
			Name:  "coinbase",
			Must:  true,
			Usage: "coinbase address hex string",
		}

		cmd := &ishell.Cmd{
			Name: "minerhistory",
			Help: "miner history reward info",
			Func: func(c *ishell.Context) {
				args := []util.Flag{coinbase}
				if util.HelpText(c, args) {
					return
				}

				coinbaseP = util.StringVar(c.Args, coinbase)

				if err := minerHistoryAction(coinbaseP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "minerhistory",
			Short: "miner history reward info",
			Run: func(cmd *cobra.Command, args []string) {
				err := minerHistoryAction(coinbaseP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&coinbaseP, "coinbase", "", "coinbase address hex string")
		rootCmd.AddCommand(cmd)
	}
}

func minerHistoryAction(coinbaseP string) error {
	if coinbaseP == "" {
		return errors.New("invalid coinbase value")
	}

	account, err := types.HexToAddress(coinbaseP)
	if err != nil {
		return err
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspHistoryInfo := api.MinerHistoryRewardInfo{}
	err = client.Call(&rspHistoryInfo, "miner_getHistoryRewardInfos", account)
	if err != nil {
		return err
	}

	gasAmount0 := new(big.Float).SetInt(rspHistoryInfo.AllRewardAmount.Int)
	gasAmount1 := new(big.Float).Quo(gasAmount0, new(big.Float).SetInt64(1e8))
	gasAmountF, _ := gasAmount1.Float64()

	fmt.Printf("AllRewardBlocks: %d\n", rspHistoryInfo.AllRewardBlocks)
	fmt.Printf("AllRewardAmount: %d (%f)\n", rspHistoryInfo.AllRewardAmount, gasAmountF)
	fmt.Printf("%-10s %-10s %-10s %-64s\n", "Start", "End", "Blocks", "Beneficial")
	for _, info := range rspHistoryInfo.RewardInfos {
		fmt.Printf("%-10d %-10d %-10d %-64s\n", info.StartHeight, info.EndHeight, info.RewardBlocks, info.Beneficial)
	}

	return nil
}
