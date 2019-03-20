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
	"github.com/qlcchain/go-qlc/cmd/client/commands"
	"github.com/spf13/cobra"
)

func version() {
	if interactive {
		c := &ishell.Cmd{
			Name: "version",
			Help: "show version info for server",
			Func: func(c *ishell.Context) {
				if commands.HelpText(c, nil) {
					return
				}
				if err := commands.CheckArgs(c, nil); err != nil {
					commands.Warn(err)
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
	v := fmt.Sprintf("%-15s%s %s", "build time:", ts[0], ts[1])
	b := fmt.Sprintf("%-15s%s", "version:", version)
	g := fmt.Sprintf("%-15s%s", "hash:", gitrev)
	var ver string
	if goqlc.MAINNET {
		ver = fmt.Sprintf("%-15s", "mainnet")
	} else {
		ver = fmt.Sprintf("%-15s", "testnet")
	}
	if interactive {
		commands.Info(ver)
		commands.Info(v)
		commands.Info(b)
		commands.Info(g)
	} else {
		fmt.Println(ver)
		fmt.Println(v)
		fmt.Println(b)
		fmt.Println(g)
	}
}
