package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addDoDSettlementCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "dod_settlement",
			Help: "dod settlement commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addDSCreateOrderCmdByShell(cmd)
		addDSCreateResponseCmdByShell(cmd)
		addDSUpdateOrderInfoCmdByShell(cmd)
		addDSUpdateResponseCmdByShell(cmd)
		addDSResourceReadyCmdByShell(cmd)
		addDSChangeOrderCmdByShell(cmd)
		addDSChangeResponseCmdByShell(cmd)
		addDSTerminateOrderCmdByShell(cmd)
		addDSTerminateResponseCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "dod_settlement",
			Short: "dod settlement commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
