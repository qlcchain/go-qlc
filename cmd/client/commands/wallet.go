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

func addWalletCmd() {
	if interactive {
		cmd := &ishell.Cmd{
			Name: "wallet",
			Help: "wallet commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(cmd)

		addWalletChangePasswordCmdByShell(cmd)
		addWalletCreateCmdByShell(cmd)
		addWalletListCmdByShell(cmd)
		addWalletRemoveCmdByShell(cmd)
	} else {
		var cmd = &cobra.Command{
			Use:   "wallet",
			Short: "wallet commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(cmd)

		addWalletChangePasswordCmdByCobra(cmd)
		addWalletCreateCmdByCobra(cmd)
		addWalletListCmdByCobra(cmd)
		addWalletRemoveCmdByCobra(cmd)
	}
}
