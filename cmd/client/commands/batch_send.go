/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
)

func init() {
	from := Flag{
		Name:  "from",
		Must:  true,
		Usage: "send account private key",
		Value: "",
	}
	to := Flag{
		Name:  "to",
		Must:  true,
		Usage: "receive accounts",
		Value: "",
	}
	token := Flag{
		Name:  "token",
		Must:  false,
		Usage: "token name for send action(defalut is QLC)",
		Value: "QLC",
	}
	amount := Flag{
		Name:  "amount",
		Must:  true,
		Usage: "send amount",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "batchsend",
		Help: "batch send transaction",
		Func: func(c *ishell.Context) {
			args := []Flag{from, to, token, amount}
			if HelpText(c, args) {
				return
			}
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			fromAccountP := StringVar(c.Args, from)
			toAccountsP := StringSliceVar(c.Args, to)
			tokenP := StringVar(c.Args, token)
			amountP := StringVar(c.Args, amount)
			if err := batchSendAction(fromAccountP, tokenP, amountP, toAccountsP); err != nil {
				Warn(err)
				return
			}
			Info("batch send done")
		},
	}
	shell.AddCmd(c)
}

func batchSendAction(fromAccountP, tokenP, amountP string, toAccountsP []string) error {
	if fromAccountP == "" || len(toAccountsP) == 0 {
		return errors.New("err transfer info")
	}
	bytes, err := hex.DecodeString(fromAccountP)
	if err != nil {
		return err
	}

	account := types.NewAccount(bytes)
	am := types.StringToBalance(amountP)

	for _, toAccount := range toAccountsP {
		t, err := types.HexToAddress(strings.TrimSpace(toAccount))
		if err != nil {
			Warn(err)
			continue
		}
		if err := sendTx(account, t, tokenP, am); err != nil {
			return err
		}
	}
	return nil
}
