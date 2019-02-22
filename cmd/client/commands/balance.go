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
				err := accountBalance(addresses)
				if err != nil {
					Warn(err)
					return
				}
			},
		}
		shell.AddCmd(s)
	} else {
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
		rootCmd.AddCommand(balanceCmd)
	}
}

func accountBalance(addresses []string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var resp map[types.Address]map[string]map[string]types.Balance
	err = client.Call(&resp, "ledger_accountsBalances", addresses)
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
				Info(a, ":")
			} else {
				fmt.Println(a, ":")
			}
			for k, v := range value {
				fmt.Printf("    %s: balance is %s, pending is %s", k, v["balance"], v["pending"])
				fmt.Println()
			}
		} else {
			if interactive {
				Info(a, " not found")
			} else {
				fmt.Println(a, " not found")
			}
		}
	}
	return nil
}
