/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

func addPledgeCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "pledge",
			Help: "pledge contract commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addPledgePledgeCmdByShell(cmd)
		addPledgeRecvPendCmdByShell(cmd)
		addPledgeWithdrawCmdByShell(cmd)
		addPledgeGetInfoCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "pledge",
			Short: "pledge contract commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)

		addPledgePledgeCmdByCobra(cmd)
		addPledgeWithdrawCmdByCobra(cmd)
	}
}
