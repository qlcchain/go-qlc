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

func addAccountCmd() {
	if interactive {
		accCmd := &ishell.Cmd{
			Name: "account",
			Help: "account commands",
			Func: func(c *ishell.Context) {
				c.Println(c.Cmd.HelpText())
			},
		}
		shell.AddCmd(accCmd)
		addGenerateAccountCmdByShell(accCmd)
	} else {
		var accCmd = &cobra.Command{
			Use:   "account",
			Short: "account commands",
			Run: func(cmd *cobra.Command, args []string) {
			},
		}
		rootCmd.AddCommand(accCmd)
		addGenerateAccountCmdByCobra(accCmd)
	}
}
