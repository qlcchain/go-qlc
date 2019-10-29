package commands

import (
	"github.com/abiosoft/ishell"

	"github.com/spf13/cobra"
)

func addLedgerCmd() {
	if interactive {
		txCmd := &ishell.Cmd{
			Name: "ledger",
			Help: "ledger commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(txCmd)
		addLedgerDumpByIshell(txCmd)
		addLedgerGCByIshell(txCmd)
	} else {
		var txCmd = &cobra.Command{
			Use:   "tx",
			Short: "tx commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(txCmd)
		addLedgerDumpByCobra(txCmd)
		addLedgerGCByCobra(txCmd)
	}
}
