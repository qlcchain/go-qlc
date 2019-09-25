package commands

import (
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/rpc/api"
	"time"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/spf13/cobra"
)

func pending() {
	var accountP string

	if interactive {
		account := util.Flag{
			Name:  "account",
			Must:  true,
			Usage: "account address string",
		}

		cmd := &ishell.Cmd{
			Name: "pending",
			Help: "get account pending info",
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

				if err := pendingAction(accountP); err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "recv",
			Short: "recv pending balance",
			Run: func(cmd *cobra.Command, args []string) {
				err := pendingAction(accountP)
				if err != nil {
					cmd.Println(err)
				}
			},
		}
		cmd.Flags().StringVar(&accountP, "account", "", "account address string")
		rootCmd.AddCommand(cmd)
	}
}

func pendingAction(accountP string) error {
	if accountP == "" {
		return errors.New("invalid address value")
	}

	account, err := types.HexToAddress(accountP)
	if err != nil {
		return errors.New("can not new address")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	allPendInfos := make(map[types.Address][]*api.APIPending)
	err = client.Call(&allPendInfos, "ledger_accountsPending", []types.Address{account}, 100)
	if err != nil {
		return err
	}

	if len(allPendInfos) == 0 {
		return nil
	}

	fmt.Printf("%-64s %-10s %-18s %-20s\n", "Hash", "Token", "Amount", "Timestamp")
	for _, addrInfos := range allPendInfos {
		for _, info := range addrInfos {
			ts := time.Unix(info.Timestamp, 0)
			fmt.Printf("%-64s %-10s %-18d %-20s\n", info.Hash, info.TokenName, info.Amount, ts.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}
