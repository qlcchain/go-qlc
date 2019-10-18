package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addTxCmd() {
	if interactive {
		txCmd := &ishell.Cmd{
			Name: "tx",
			Help: "tx commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(txCmd)

		addTxBlockInfoCmdByShell(txCmd)
		addTxBlockListCmdByShell(txCmd)
		addTxPendingCmdByShell(txCmd)
		addTxChangeCmdByShell(txCmd)
		addTxRecvCmdByShell(txCmd)
		addTxSendCmdByShell(txCmd)
	} else {
		var txCmd = &cobra.Command{
			Use:   "tx",
			Short: "tx commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(txCmd)

		addTxPendingCmdByCobra(txCmd)
		addTxChangeCmdByCobra(txCmd)
		addTxRecvCmdByCobra(txCmd)
		addTxSendCmdByCobra(txCmd)
	}
}
