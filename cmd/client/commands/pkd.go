package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

// public key distribution cmd
func addPKDCmd() {
	if interactive {
		PKDCmd := &ishell.Cmd{
			Name: "pkd",
			Help: "public key distribution commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(PKDCmd)

		addVerifierRegisterCmdByShell(PKDCmd)
		addVerifierUnRegisterCmdByShell(PKDCmd)
		addPublishCmdByShell(PKDCmd)
		addUnPublishCmdByShell(PKDCmd)
		addOraclePublishCmdByShell(PKDCmd)
	} else {
		var PKDCmd = &cobra.Command{
			Use:   "pkd",
			Short: "public key distribution commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(PKDCmd)
	}
}
