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

func walletCreate() {
	var pwdP string
	if interactive {
		pwd := Flag{
			Name:  "password",
			Must:  true,
			Usage: "password for new wallet",
			Value: "",
		}
		c := &ishell.Cmd{
			Name: "walletcreate",
			Help: "create a wallet for QLCChain node",
			Func: func(c *ishell.Context) {
				args := []Flag{pwd}
				if HelpText(c, args) {
					return
				}
				if err := CheckArgs(c, args); err != nil {
					Warn(err)
					return
				}
				pwdP = StringVar(c.Args, pwd)
				addr, err := createWallet(pwdP)
				if err != nil {
					Warn(err)
				} else {
					Info(fmt.Sprintf("create wallet: address=>%s, password=>%s success", addr.String(), pwdP))
				}
			},
		}
		shell.AddCmd(c)
	} else {
		var wcCmd = &cobra.Command{
			Use:   "walletcreate",
			Short: "create a wallet for QLCChain node",
			Run: func(cmd *cobra.Command, args []string) {
				addr, err := createWallet(pwdP)
				if err != nil {
					cmd.Println(err)
				} else {
					cmd.Printf("create wallet: address=>%s, password=>%s success", addr.String(), pwdP)
					cmd.Println()
				}
			},
		}
		wcCmd.Flags().StringVarP(&pwdP, "password", "p", "", "password for wallet")
		rootCmd.AddCommand(wcCmd)
	}
}

func createWallet(pwdP string) (types.Address, error) {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return types.ZeroAddress, err
	}
	defer client.Close()
	var addr types.Address
	err = client.Call(&addr, "wallet_newWallet", pwdP)
	if err != nil {
		return types.ZeroAddress, err
	}
	return addr, nil
}
