/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/chain/version"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/spf13/cobra"
)

func chainVersion() {
	if interactive {
		c := &ishell.Cmd{
			Name: "version",
			Help: "show version info for server",
			Func: func(c *ishell.Context) {
				if util.HelpText(c, nil) {
					return
				}
				if err := util.CheckArgs(c, nil); err != nil {
					util.Warn(err)
					return
				}
				versionInfo()
			},
		}
		shell.AddCmd(c)
	} else {
		var versionCmd = &cobra.Command{
			Use:   "version",
			Short: "show version info",
			Run: func(cmd *cobra.Command, args []string) {
				versionInfo()
			},
		}
		rootCmd.AddCommand(versionCmd)
	}
}

func versionInfo() {
	v := version.VersionString()
	if interactive {
		util.Info(v)
	} else {
		fmt.Println(v)
	}
}
