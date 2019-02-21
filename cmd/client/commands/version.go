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

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc"
	"github.com/spf13/cobra"
)

func version() {
	if interactive {
		c := &ishell.Cmd{
			Name: "version",
			Help: "show version info for client",
			Func: func(c *ishell.Context) {
				if HelpText(c, nil) {
					return
				}
				if err := CheckArgs(c, nil); err != nil {
					Warn(err)
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
	version := goqlc.VERSION
	buildTime := goqlc.BUILDTIME
	gitrev := goqlc.GITREV
	ts := strings.Split(buildTime, "_")
	if interactive {
		Info(fmt.Sprintf("%-15s%s %s", "build time:", ts[0], ts[1]))
		Info(fmt.Sprintf("%-15s%s", "version:", version))
		Info(fmt.Sprintf("%-15s%s", "hash:", gitrev))
	} else {
		fmt.Println(fmt.Sprintf("%-15s%s %s", "build time:", ts[0], ts[1]))
		fmt.Println(fmt.Sprintf("%-15s%s", "version:", version))
		fmt.Println(fmt.Sprintf("%-15s%s", "hash:", gitrev))
	}
}
