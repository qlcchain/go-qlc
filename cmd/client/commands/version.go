/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"strings"

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"
)

var (
	Version   = ""
	GitRev    = ""
	BuildTime = ""
)

func version() {
	if interactive {
		c := &ishell.Cmd{
			Name: "version",
			Help: "show version info for client",
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
	ts := strings.Split(BuildTime, "_")
	if interactive {
		util.Info(fmt.Sprintf("%-15s%s %s", "build time:", ts[0], ts[1]))
		util.Info(fmt.Sprintf("%-15s%s", "version:", Version))
		util.Info(fmt.Sprintf("%-15s%s", "hash:", GitRev))
	} else {
		fmt.Println(fmt.Sprintf("%-15s%s %s", "build time:", ts[0], ts[1]))
		fmt.Println(fmt.Sprintf("%-15s%s", "version:", Version))
		fmt.Println(fmt.Sprintf("%-15s%s", "hash:", GitRev))
	}
}
