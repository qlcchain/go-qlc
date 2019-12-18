/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/common/types"
)

func addWalletChangePasswordCmdByShell(parentCmd *ishell.Cmd) {
	account := util.Flag{
		Name:  "account",
		Must:  true,
		Usage: "account for wallet",
		Value: "",
	}
	pwd := util.Flag{
		Name:  "password",
		Must:  true,
		Usage: "password for wallet",
		Value: "",
	}
	newPwd := util.Flag{
		Name:  "newpassword",
		Must:  true,
		Usage: "new password for wallet",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "changepassword",
		Help: "change wallet password",
		Func: func(c *ishell.Context) {
			args := []util.Flag{account, pwd, newPwd}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			accountP := util.StringVar(c.Args, account)
			passwordP := util.StringVar(c.Args, pwd)
			newpasswordP := util.StringVar(c.Args, newPwd)
			err := changePwd(accountP, passwordP, newpasswordP)
			if err != nil {
				util.Warn(err)
			} else {
				util.Info(fmt.Sprintf("change password success for account: %s", accountP))
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addWalletChangePasswordCmdByCobra(parentCmd *cobra.Command) {
	var accountP string
	var passwordP string
	var newpasswordP string
	var wcpCmd = &cobra.Command{
		Use:   "changepassword",
		Short: "change wallet password",
		Run: func(cmd *cobra.Command, args []string) {
			err := changePwd(accountP, passwordP, newpasswordP)
			if err != nil {
				cmd.Println(err)
			} else {
				cmd.Printf("change password success for account: %s", accountP)
				cmd.Println()
				return
			}
		},
	}
	wcpCmd.Flags().StringVarP(&accountP, "account", "a", "", "wallet address")
	wcpCmd.Flags().StringVarP(&passwordP, "password", "p", "", "password for wallet")
	wcpCmd.Flags().StringVarP(&newpasswordP, "newpassword", "n", "", "new password for wallet")
	parentCmd.AddCommand(wcpCmd)
}

func changePwd(accountP, pwdP, newPwdP string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()
	var addr types.Address
	err = client.Call(&addr, "wallet_changePassword", accountP, pwdP, newPwdP)
	if err != nil {
		return err
	}
	return nil
}
