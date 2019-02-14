/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"strings"

	"github.com/qlcchain/go-qlc"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version info",
	Run: func(cmd *cobra.Command, args []string) {
		version := goqlc.VERSION
		buildTime := goqlc.BUILDTIME
		gitrev := goqlc.GITREV
		ts := strings.Split(buildTime, "_")
		cmd.Printf("build time: %s %s ", ts[0], ts[1])
		cmd.Println()
		cmd.Println("version:   ", version)
		cmd.Println("hash: 	   ", gitrev)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
