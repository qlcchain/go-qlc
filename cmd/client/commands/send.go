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
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/qlcchain/go-qlc/rpc/api"
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
		Usage: "receive account",
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
		Name: "send",
		Help: "send transaction",
		Func: func(c *ishell.Context) {
			args := []Flag{from, to, token, amount}
			if HelpText(c, args) {
				return
			}
			if err := CheckArgs(c, args); err != nil {
				Warn(err)
				return
			}
			fromP := StringVar(c.Args, from)
			toP := StringVar(c.Args, to)
			tokenP := StringVar(c.Args, token)
			amountP := StringVar(c.Args, amount)
			err := sendAction(fromP, toP, tokenP, amountP)
			if err != nil {
				Warn(err)
				return
			}
			Info("send transaction success!")
		},
	}
	shell.AddCmd(c)
}

func sendAction(fromP, toP, tokenP, amountP string) error {
	if fromP == "" || toP == "" || amountP == "" {
		fmt.Println("err transfer info")
		return errors.New("err transfer info")
	}
	bytes, err := hex.DecodeString(fromP)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fromAccount := types.NewAccount(bytes)

	t, err := types.HexToAddress(toP)
	if err != nil {
		fmt.Println(err)
		return err
	}

	am := types.StringToBalance(amountP)
	if err := sendTx(fromAccount, t, tokenP, am); err != nil {
		return err
	}
	return nil
}

func sendTx(account *types.Account, to types.Address, token string, amount types.Balance) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	//a, _ := mock.BalanceToRaw(amount, "QLC")
	//fmt.Println(a)
	para := api.APISendBlockPara{
		Send:      account.Address(),
		TokenName: token,
		To:        to,
		Amount:    amount,
	}
	var sendBlock types.StateBlock
	err = client.Call(&sendBlock, "ledger_generateSendBlock", para, hex.EncodeToString(account.PrivateKey()))
	if err != nil {
		return err
	}
	from := account.Address()
	Info(fmt.Sprintf("send token(%s) from %s  to %s", token, from.String(), to.String()))
	//Info(fmt.Sprintf("block hash: %s", sendBlock.GetHash()))

	var h types.Hash
	err = client.Call(&h, "ledger_process", &sendBlock)
	if err != nil {
		Warn("process block error, ", err)
		return err
	}
	return nil
}
