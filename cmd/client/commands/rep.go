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

func addRepCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "rep",
			Help: "rep commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addRepRewardCmdByShell(cmd)
		addRepRewardRecvpendCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "rep",
			Short: "rep commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)

		addRepRewardCmdByCobra(cmd)
		addRepRewardRecvpendByCobra(cmd)
	}
}
