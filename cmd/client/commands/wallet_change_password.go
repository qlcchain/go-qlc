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
)

func init() {
	account := Flag{
		Name:  "account",
		Must:  true,
		Usage: "account for wallet",
		Value: "",
	}
	pwd := Flag{
		Name:  "password",
		Must:  true,
		Usage: "password for wallet",
		Value: "",
	}
	newPwd := Flag{
		Name:  "newpassword",
		Must:  true,
		Usage: "new password for wallet",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "changepassword",
		Help: "change wallet password",
		Func: func(c *ishell.Context) {
			args := []Flag{account, pwd, newPwd}
			if HelpText(c, args) {
				return
			}
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			accountP := StringVar(c.Args, account)
			pwdP := StringVar(c.Args, pwd)
			newPwdP := StringVar(c.Args, newPwd)
			err := changePassword(accountP, pwdP, newPwdP)
			if err != nil {
				Warn(err)
			} else {
				Info(fmt.Sprintf("change password success for account: %s", accountP))
			}
		},
	}
	shell.AddCmd(c)

}

func changePassword(accountP, pwdP, newPwdP string) error {
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
