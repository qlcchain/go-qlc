package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addPtmKeyCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "ptmkey",
			Help: "ptmkey update commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addPtmKeyUpdateCmdByShell(cmd)
		addPtmKeyGetPubkeyCmdByShell(cmd)
		addPtmKeyDeleteByAccountByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "ptmkey",
			Short: "ptmkey update commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
