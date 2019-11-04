package commands

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"

	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
)

func addMinerRecvPendCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account private hex string",
	}
	sendHash := util.Flag{
		Name:  "hash",
		Must:  true,
		Usage: "reward send block hash string",
	}

	cmd := &ishell.Cmd{
		Name: "recvPending",
		Help: "miner recv pending reward (gas token)",
		Func: func(c *ishell.Context) {
			args := []util.Flag{account, sendHash}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			accountP := util.StringVar(c.Args, account)
			sendHashP := util.StringVar(c.Args, sendHash)

			if err := minerRecvPendAction(accountP, sendHashP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func addMinerRecvPendCmdByCobra(parentCmd *cobra.Command) {
	var accountP string
	var sendHashP string

	var cmd = &cobra.Command{
		Use:   "recvPending",
		Short: "miner recv pending reward (gas token)",
		Run: func(cmd *cobra.Command, args []string) {
			err := minerRecvPendAction(accountP, sendHashP)
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&accountP, "account", "", "account private hex string")
	cmd.Flags().StringVar(&sendHashP, "hash", "", "reward send block hash string")
	parentCmd.AddCommand(cmd)
}

func minerRecvPendAction(accountP, sendHashP string) error {
	if accountP == "" {
		return errors.New("invalid account value")
	}

	if sendHashP == "" {
		return errors.New("invalid hash value")
	}

	accBytes, err := hex.DecodeString(accountP)
	if err != nil {
		return err
	}
	account := types.NewAccount(accBytes)
	if account == nil {
		return errors.New("can not new account")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	reward := types.StateBlock{}
	err = client.Call(&reward, "miner_getRewardRecvBlockBySendHash", sendHashP)
	if err != nil {
		return err
	}

	var w2 types.Work
	worker2, _ := types.NewWorker(w2, reward.Root())
	reward.Work = worker2.NewWork()

	rewardHash := reward.GetHash()
	reward.Signature = account.Sign(rewardHash)

	fmt.Printf("RewardBlock:\n%s\n", cutil.ToIndentString(reward))
	fmt.Println("address", reward.Address, "rewardHash", rewardHash)

	err = client.Call(nil, "ledger_process", &reward)
	if err != nil {
		return err
	}

	fmt.Printf("success to recv reward, account balance %s\n", reward.Balance)

	return nil
}
