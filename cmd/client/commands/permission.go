package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addPermissionCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "permission",
			Help: "permission control commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addPermissionAdminUpdateSendCmdByShell(cmd)
		addPermissionAdminUpdateRewardCmdByShell(cmd)
		addPermissionNodeAddCmdByShell(cmd)
		addPermissionNodeUpdateCmdByShell(cmd)
		addPermissionNodeRemoveCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "permission",
			Short: "permission control commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
