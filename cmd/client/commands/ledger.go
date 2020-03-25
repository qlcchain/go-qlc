package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addLedgerCmd() {
	if interactive {
		ledgerCmd := &ishell.Cmd{
			Name: "ledger",
			Help: "ledger commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(ledgerCmd)
		addLedgerDumpByIshell(ledgerCmd)
		addLedgerGCByIshell(ledgerCmd)
		addLedgerGenerateTestLedgerByIshell(ledgerCmd)
		addLedgerBlockCountByIshell(ledgerCmd)
		addLedgerTokensByIshell(ledgerCmd)
		addLedgerBalanceByIshell(ledgerCmd)
	} else {
		var ledgerCmd = &cobra.Command{
			Use:   "ledger",
			Short: "ledger commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(ledgerCmd)
		addLedgerDumpByCobra(ledgerCmd)
		addLedgerGCByCobra(ledgerCmd)
		addLedgerGenerateTestLedgerByCobra(ledgerCmd)
		addLedgerBlockCountByCobra(ledgerCmd)
		addLedgerBalanceByCobra(ledgerCmd)
	}
}
