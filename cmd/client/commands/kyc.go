package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addKYCCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "kyc",
			Help: "kyc status commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addKYCAdminUpdateCmdByShell(cmd)
		addKYCStatusUpdateCmdByShell(cmd)
		addKYCTradeAddressUpdateCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "kyc",
			Short: "kyc status commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)
	}
}
