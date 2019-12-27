package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addOracleCmd() {
	if interactive {
		OracleCmd := &ishell.Cmd{
			Name: "oracle",
			Help: "oracle commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(OracleCmd)

		addOraclePublishCmdByShell(OracleCmd)
	} else {
		var OracleCmd = &cobra.Command{
			Use:   "oracle",
			Short: "oracle commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(OracleCmd)
	}
}
