package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addVerifierCmd() {
	if interactive {
		verifierCmd := &ishell.Cmd{
			Name: "verifier",
			Help: "verifier commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(verifierCmd)

		addVerifierRegisterCmdByShell(verifierCmd)
		addVerifierUnRegisterCmdByShell(verifierCmd)
	} else {
		var dbgCmd = &cobra.Command{
			Use:   "verifier",
			Short: "verifier commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(dbgCmd)
	}
}
