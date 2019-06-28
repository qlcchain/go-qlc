package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc/api"
	"github.com/spf13/cobra"
)

func minerReward() {
	var coinbaseP string
	var beneficialP string
	var heightP int

	if interactive {
		coinbase := util.Flag{
			Name:  "coinbase",
			Must:  true,
			Usage: "coinbase coinbase private hex string",
		}
		beneficial := util.Flag{
			Name:  "beneficial",
			Must:  true,
			Usage: "beneficial coinbase private hex string",
		}
		height := util.Flag{
			Name:  "height",
			Must:  true,
			Usage: "reward height",
		}

		cmd := &ishell.Cmd{
			Name: "minerreward",
			Help: "miner get reward (gas token)",
			Func: func(c *ishell.Context) {
				args := []util.Flag{coinbase}
				if util.HelpText(c, args) {
					return
				}
				err := util.CheckArgs(c, args)
				if err != nil {
					util.Warn(err)
					return
				}

				coinbaseP = util.StringVar(c.Args, coinbase)
				beneficialP = util.StringVar(c.Args, beneficial)
				heightP, _ = util.IntVar(c.Args, height)

				if err := minerRewardAction(coinbaseP, beneficialP, heightP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "minerreward",
			Short: "miner get reward (gas token)",
			Run: func(cmd *cobra.Command, args []string) {
				err := minerRewardAction(coinbaseP, beneficialP, heightP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&coinbaseP, "coinbase", "", "coinbase account private hex string")
		cmd.Flags().StringVar(&beneficialP, "beneficial", "", "beneficial account private hex string")
		cmd.Flags().IntVar(&heightP, "height", 0, "reward height")
		rootCmd.AddCommand(cmd)
	}
}

func minerRewardAction(coinbaseP, beneficialP string, heightP int) error {
	if coinbaseP == "" {
		return errors.New("invalid coinbase value")
	}

	if beneficialP == "" {
		return errors.New("invalid beneficial value")
	}

	if heightP <= 0 {
		return errors.New("invalid height value")
	}

	cbBytes, err := hex.DecodeString(coinbaseP)
	if err != nil {
		return err
	}
	coinbaseAcc := types.NewAccount(cbBytes)
	if coinbaseAcc == nil {
		return errors.New("can not new coinbase account")
	}

	benBytes, err := hex.DecodeString(beneficialP)
	if err != nil {
		return err
	}
	beneficialAcc := types.NewAccount(benBytes)
	if beneficialAcc == nil {
		return errors.New("can not new beneficial account")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rewardParam := api.RewardParam{
		Coinbase:     coinbaseAcc.Address(),
		Beneficial:   beneficialAcc.Address(),
		RewardHeight: uint64(heightP),
	}

	send := types.StateBlock{}
	err = client.Call(&send, "miner_getRewardSendBlock", &rewardParam)
	if err != nil {
		return err
	}

	sendHash := send.GetHash()
	send.Signature = coinbaseAcc.Sign(sendHash)
	var w types.Work
	worker, _ := types.NewWorker(w, send.Root())
	send.Work = worker.NewWork()

	fmt.Println("address", send.Address, "sendHash", sendHash)

	reward := types.StateBlock{}
	err = client.Call(&reward, "miner_getRewardRecvBlock", &send)
	if err != nil {
		return err
	}

	rewardHash := reward.GetHash()
	reward.Signature = beneficialAcc.Sign(rewardHash)
	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	fmt.Println("address", reward.Address, "rewardHash", rewardHash)

	err = client.Call(nil, "ledger_process", &send)
	if err != nil {
		return err
	}

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		return err
	}

	fmt.Println("success to get miner reward, please check account balance")

	return nil
}
