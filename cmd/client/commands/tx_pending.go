package commands

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/rpc/api"
	"time"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/spf13/cobra"
)

func addTxPendingCmdByShell(parentCmd *ishell.Cmd) {
	addressFlag := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "account address string",
	}

	cmd := &ishell.Cmd{
		Name: "getPending",
		Help: "get account pending info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{addressFlag}
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			address := util.StringVar(c.Args, addressFlag)

			if err := pendingAction(address); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func addTxPendingCmdByCobra(parentCmd *cobra.Command) {
	var accountP string

	var cmd = &cobra.Command{
		Use:   "getPending",
		Short: "get account pending info",
		Run: func(cmd *cobra.Command, args []string) {
			err := pendingAction(accountP)
			if err != nil {
				cmd.Println(err)
			}
		},
	}
	cmd.Flags().StringVar(&accountP, "address", "", "account address string")
	parentCmd.AddCommand(cmd)
}

func pendingAction(address string) error {
	if address == "" {
		return errors.New("invalid address value")
	}

	account, err := types.HexToAddress(address)
	if err != nil {
		return errors.New("can not new address")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	allPendInfos := make(map[types.Address][]*api.APIPending)
	err = client.Call(&allPendInfos, "ledger_accountsPending", []types.Address{account}, 1000)
	if err != nil {
		return err
	}

	if len(allPendInfos) == 0 {
		return nil
	}

	pendCnt := 0
	for _, addrInfos := range allPendInfos {
		pendCnt += len(addrInfos)
	}

	fmt.Printf("Pendings Count: %d\n", pendCnt)

	fmt.Printf("%-64s %-15s %-10s %-18s %-20s\n", "Hash", "Type", "Token", "Amount", "Timestamp")
	for _, addrInfos := range allPendInfos {
		for _, info := range addrInfos {
			ts := time.Unix(info.Timestamp, 0)
			fmt.Printf("%-64s %-15s %-10s %-18d %-20s\n",
				info.Hash, info.BlockType,
				info.TokenName, info.Amount,
				ts.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}
