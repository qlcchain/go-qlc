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
)

func init() {
	c := &ishell.Cmd{
		Name: "version",
		Help: "show version info for client",
		Func: func(c *ishell.Context) {
			if HelpText(c, nil) {
				return
			}
			version := goqlc.VERSION
			buildTime := goqlc.BUILDTIME
			gitrev := goqlc.GITREV
			ts := strings.Split(buildTime, "_")
			Info(fmt.Sprintf("build time: %s %s ", ts[0], ts[1]))
			Info("version:   ", version)
			Info("hash:      ", gitrev)
		},
	}
	shell.AddCmd(c)
}
