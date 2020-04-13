package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addPrivacyCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "privacy",
			Help: "privacy contract commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addPrivacyDistributePayloadCmdByShell(cmd)
		addPrivacyGetPayloadCmdByShell(cmd)

		addPrivacyDemoKVSetCmdByShell(cmd)
		addPrivacyDemoKVGetCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "privacy",
			Short: "privacy contract commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
