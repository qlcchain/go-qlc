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
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addWalletRemoveCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account for wallet",
		Value: "",
	}
	args := []util.Flag{account}
	c := &ishell.Cmd{
		Name:                "remove",
		Help:                "remove a wallet",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			accountP := util.StringVar(c.Args, account)

			err := removeWallet(accountP)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addWalletRemoveCmdByCobra(parentCmd *cobra.Command) {
	var accountP string
	var wrCmd = &cobra.Command{
		Use:   "remove",
		Short: "remove wallet",
		Run: func(cmd *cobra.Command, args []string) {
			err := removeWallet(accountP)
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	wrCmd.Flags().StringVarP(&accountP, "account", "a", "", "wallet address")
	parentCmd.AddCommand(wrCmd)
}

func removeWallet(accountP string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	err = client.Call(nil, "wallet_remove", accountP)
	if err != nil {
		return err
	}
	s := fmt.Sprintf("remove wallet %s success", accountP)
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}
	return nil
}
