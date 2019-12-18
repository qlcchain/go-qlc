/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addLedgerBalanceByIshell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "address for account",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "balance",
		Help: "balance for accounts",
		Func: func(c *ishell.Context) {
			args := []util.Flag{address}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			addresses := util.StringSliceVar(c.Args, address)
			err := accountBalance(addresses)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addLedgerBalanceByCobra(parentCmd *cobra.Command) {
	var addresses []string
	var balanceCmd = &cobra.Command{
		Use:   "balance",
		Short: " balance for accounts",
		Run: func(cmd *cobra.Command, args []string) {
			if len(addresses) < 1 {
				cmd.Println("err account")
				return
			}
			err := accountBalance(addresses)
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	balanceCmd.Flags().StringSliceVar(&addresses, "address", addresses, "address for accounts")
	parentCmd.AddCommand(balanceCmd)
}

func accountBalance(addresses []string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var resp map[types.Address]map[string]map[string]types.Balance
	err = client.Call(&resp, "ledger_accountsBalance", addresses)
	if err != nil {
		return err
	}

	for _, a := range addresses {
		addr, err := types.HexToAddress(a)
		if err != nil {
			return err
		}
		if value, ok := resp[addr]; ok {
			if interactive {
				util.Info(a, ":")
			} else {
				fmt.Println(a, ":")
			}
			for k, v := range value {
				fmt.Printf("    %s: balance is %s, pending is %s", k, v["balance"], v["pending"])
				fmt.Println()
			}
		} else {
			if interactive {
				util.Info(a, " not found")
			} else {
				fmt.Println(a, " not found")
			}
		}
	}
	return nil
}
