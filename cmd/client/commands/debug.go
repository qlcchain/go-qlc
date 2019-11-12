package commands

import (
	"github.com/abiosoft/ishell"

	"github.com/spf13/cobra"
)

func addDebugCmd() {
	if interactive {
		dbgCmd := &ishell.Cmd{
			Name: "debug",
			Help: "debug commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(dbgCmd)

		addDebugPovInfoCmdByShell(dbgCmd)
		addDebugConsensusInfoCmdByShell(dbgCmd)
	} else {
		var dbgCmd = &cobra.Command{
			Use:   "debug",
			Short: "debug commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(dbgCmd)
	}
}
