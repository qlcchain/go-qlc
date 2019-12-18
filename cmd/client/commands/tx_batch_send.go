package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addTxBatchSendByShell(parentCmd *ishell.Cmd) {
	from := util.Flag{
		Name:  "from",
		Must:  true,
		Usage: "send account private key",
		Value: "",
	}
	to := util.Flag{
		Name:  "to",
		Must:  true,
		Usage: "receive accounts",
		Value: "",
	}
	token := util.Flag{
		Name:  "token",
		Must:  false,
		Usage: "token name for send action(defalut is QLC)",
		Value: "QLC",
	}
	amount := util.Flag{
		Name:  "amount",
		Must:  true,
		Usage: "send amount",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "batchsend",
		Help: "batch send transaction",
		Func: func(c *ishell.Context) {
			args := []util.Flag{from, to, token, amount}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}
			fromAccountP := util.StringVar(c.Args, from)
			toAccountsP := util.StringSliceVar(c.Args, to)
			tokenP := util.StringVar(c.Args, token)
			amountP := util.StringVar(c.Args, amount)

			for _, toAccount := range toAccountsP {
				if err := sendAction(fromAccountP, toAccount, tokenP, amountP); err != nil {
					util.Warn(err)
					return
				}
			}
			util.Info("batch transaction done")
		},
	}
	parentCmd.AddCmd(c)
}

func addTxBatchSendByCobra(parentCmd *cobra.Command) {
	var fromAccountP string
	var toAccountsP []string
	var tokenP string
	var amountP string
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
	parentCmd.AddCommand(batchSendCmd)
}
