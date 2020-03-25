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
	"github.com/qlcchain/go-qlc/common/types"
)

func addWalletCreateCmdByShell(parentCmd *ishell.Cmd) {
	pwd := util.Flag{
		Name:  "password",
		Must:  false,
		Usage: "password for new wallet",
		Value: "",
	}
	seed := util.Flag{
		Name:  "seed",
		Must:  false,
		Usage: "seed for wallet",
		Value: "",
	}
	args := []util.Flag{pwd, seed}
	c := &ishell.Cmd{
		Name:                "create",
		Help:                "create a wallet for QLCChain node",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			pwdP := util.StringVar(c.Args, pwd)
			seedP := util.StringVar(c.Args, seed)
			err := createWallet(pwdP, seedP)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addWalletCreateCmdByCobra(parentCmd *cobra.Command) {
	var pwdP string
	var seedP string
	var wcCmd = &cobra.Command{
		Use:   "create",
		Short: "create a wallet for QLCChain node",
		Run: func(cmd *cobra.Command, args []string) {
			err := createWallet(pwdP, seedP)
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	wcCmd.Flags().StringVarP(&seedP, "seed", "s", "", "seed for wallet")
	wcCmd.Flags().StringVarP(&pwdP, "password", "p", "", "password for wallet")
	parentCmd.AddCommand(wcCmd)
}

func createWallet(pwdP, seedP string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var addr types.Address
	if seedP == "" {
		err = client.Call(&addr, "wallet_newWallet", pwdP)
	} else {
		err = client.Call(&addr, "wallet_newWallet", pwdP, seedP)
	}
	if err != nil {
		return err
	}
	s := fmt.Sprintf("create wallet: address=>%s, password=>%s success", addr.String(), pwdP)
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}
	return nil
}
