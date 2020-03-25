/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
)

func addSendToCreateByShell(parentCmd *ishell.Cmd) {
	from := util.Flag{
		Name:  "from",
		Must:  true,
		Usage: "send account private key",
		Value: "",
	}
	token := util.Flag{
		Name:  "token",
		Must:  false,
		Usage: "token name for send action(default is QLC)",
		Value: "QLC",
	}
	amount := util.Flag{
		Name:  "amount",
		Must:  true,
		Usage: "send amount",
		Value: "",
	}
	size := util.Flag{
		Name:  "size",
		Must:  false,
		Usage: "to be created account size",
		Value: "1",
	}
	c := &ishell.Cmd{
		Name: "sendToCreate",
		Help: "generate account(s) by Tx",
		Func: func(c *ishell.Context) {
			args := []util.Flag{from, token, amount, size}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			fromAccountP := util.StringVar(c.Args, from)
			tokenP := util.StringVar(c.Args, token)
			amountP := util.StringVar(c.Args, amount)
			accountSizeP, err := util.IntVar(c.Args, size)
			if err != nil {
				util.Warn(err)
				return
			}
			sendToCreatAction(fromAccountP, tokenP, amountP, accountSizeP, newShellPrinter())
			util.Info("sendToCreate done")
		},
	}
	parentCmd.AddCmd(c)
}

func addSendToCreateByCobra(parentCmd *cobra.Command) {
	var (
		fromAccountP string
		tokenP       string
		amountP      string
		accountSizeP int
	)

	var sendToCreateCmd = &cobra.Command{
		Use:   "sendToCreate",
		Short: "generate account(s) by Tx",
		Run: func(cmd *cobra.Command, args []string) {
			sendToCreatAction(fromAccountP, tokenP, amountP, accountSizeP, newCmdPrinter(cmd))
			cmd.Println("sendToCreate done")
		},
	}
	sendToCreateCmd.Flags().StringVar(&fromAccountP, "from", "", "send account private key")
	sendToCreateCmd.Flags().StringVar(&tokenP, "token", "QLC", "token name for send action")
	sendToCreateCmd.Flags().StringVar(&amountP, "amount", "", "send amount")
	sendToCreateCmd.Flags().IntVarP(&accountSizeP, "size", "", 1, "to be created account size")
	parentCmd.AddCommand(sendToCreateCmd)
}

func sendToCreatAction(fromAccount, token, amount string, accountSize int, log printer) {
	for i := 0; i < accountSize; i++ {
		seed, _ := types.NewSeed()
		toAccount, _ := seed.Account(0)
		if err := sendAction(fromAccount, toAccount.Address().String(), token, amount); err != nil {
			log.Warn(err)
		} else {
			log.Infof("seed: %s, account: %s", seed.String(), toAccount.String())
		}
	}
}
