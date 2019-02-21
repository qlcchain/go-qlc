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
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

func balance() {
	var addresses []string
	if interactive {
		address := Flag{
			Name:  "address",
			Must:  true,
			Usage: "address for account",
			Value: "",
		}
		s := &ishell.Cmd{
			Name: "balance",
			Help: "balance for accounts",
			Func: func(c *ishell.Context) {
				args := []Flag{address}
				if HelpText(c, args) {
					return
				}
				if err := CheckArgs(c, args); err != nil {
					Warn(err)
					return
				}
				addresses = StringSliceVar(c.Args, address)

				resp, err := accountBalance(addresses)
				if err != nil {
					Warn(err)
					return
				}
				for key, value := range resp {
					Info(key, ":")
					for k, v := range value {
						fmt.Printf("    %s: balance is %s, pending is %s", k, v["balance"], v["pending"])
						fmt.Println()
					}
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var balanceCmd = &cobra.Command{
			Use:   "balance",
			Short: " balance for accounts",
			Run: func(cmd *cobra.Command, args []string) {
				client, err := rpc.Dial(endpointP)
				if err != nil {
					cmd.Println(err)
					return
				}
				defer client.Close()
				if len(addresses) < 1 {
					cmd.Println("err account")
					return
				}

				resp, err := accountBalance(addresses)
				if err != nil {
					return
				}
				for key, value := range resp {
					cmd.Println(key)
					for k, v := range value {
						cmd.Printf("	%s, balance:%s, pending:%s", k, v["balance"], v["pending"])
						cmd.Println()
					}
				}
			},
		}
		balanceCmd.Flags().StringSliceVar(&addresses, "address", addresses, "address for accounts")
		rootCmd.AddCommand(balanceCmd)
	}
}

func accountBalance(addresses []string) (map[types.Address]map[string]map[string]types.Balance, error) {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	var resp map[types.Address]map[string]map[string]types.Balance
	err = client.Call(&resp, "ledger_accountsBalances", addresses)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
