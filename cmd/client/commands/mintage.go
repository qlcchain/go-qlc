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

func addMintageCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "mintage",
			Help: "mintage commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addMintageMintageCmdByShell(cmd)
		addMintageWithdrawCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "mintage",
			Short: "mintage commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)

		addMintageMintageCmdByCobra(cmd)
		addMintageWithdrawCmdByCobra(cmd)
	}
}
