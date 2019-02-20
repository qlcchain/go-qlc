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
	"github.com/qlcchain/go-qlc/rpc"
)

func init() {
	account := Flag{
		Name:  "account",
		Must:  true,
		Usage: "account for wallet",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "walletremove",
		Help: "remove a wallet",
		Func: func(c *ishell.Context) {
			args := []Flag{account}
			if HelpText(c, args) {
				return
			}
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			accountP := StringVar(c.Args, account)

			err := removeWallet(accountP)
			if err != nil {
				Warn(err)
			} else {
				Info(fmt.Sprintf("remove wallet: %s success", accountP))
			}
		},
	}
	shell.AddCmd(c)
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
	return nil
}
