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
)

func batchSend() {
	var fromAccountP string
	var toAccountsP []string
	var tokenP string
	var amountP string

	if interactive {
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
				fromAccountP = StringVar(c.Args, from)
				toAccountsP = StringSliceVar(c.Args, to)
				tokenP := StringVar(c.Args, token)
				amountP := StringVar(c.Args, amount)

				for _, toAccount := range toAccountsP {
					if err := sendAction(fromAccountP, toAccount, tokenP, amountP); err != nil {
						Warn(err)
						return
					}
				}
				Info("batch transaction done")
			},
		}
		shell.AddCmd(c)
	} else {
		var batchSendCmd = &cobra.Command{
			Use:   "batchsend",
			Short: "batch send transaction",
			Run: func(cmd *cobra.Command, args []string) {
				for _, toAccount := range toAccountsP {
					if err := sendAction(fromAccountP, toAccount, tokenP, amountP); err != nil {
						cmd.Println(err)
						return
					}
				}
				cmd.Println("batch transaction done")
			},
		}
		batchSendCmd.Flags().StringVar(&fromAccountP, "from", "", "send account private key")
		batchSendCmd.Flags().StringSliceVar(&toAccountsP, "to", []string{}, "receive accounts")
		batchSendCmd.Flags().StringVar(&tokenP, "token", "QLC", "token name for send action")
		batchSendCmd.Flags().StringVar(&amountP, "amount", "", "send amount")
		rootCmd.AddCommand(batchSendCmd)
	}
}
