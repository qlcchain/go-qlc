package commands

import (
	"github.com/abiosoft/ishell"

	"github.com/spf13/cobra"
)

func addCommonCmd() {
	if interactive {
		cmCmd := &ishell.Cmd{
			Name: "common",
			Help: "common commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmCmd)
		addGenerateAccountCmdByShell(cmCmd)
	} else {
		var cmCmd = &cobra.Command{
			Use:   "common",
			Short: "common commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		addGenerateAccountCmdByCobra(cmCmd)
		rootCmd.AddCommand(cmCmd)
	}
}
