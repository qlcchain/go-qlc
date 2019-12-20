/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
)

type printer interface {
	Info(a ...interface{})
	Infof(format string, i ...interface{})
	Warn(a ...interface{})
	Warnf(format string, i ...interface{})
}

type cmdPrinter struct {
	cmd *cobra.Command
}

func newCmdPrinter(cmd *cobra.Command) *cmdPrinter {
	return &cmdPrinter{cmd: cmd}
}

func (c *cmdPrinter) Info(a ...interface{}) {
	c.cmd.Println(a...)
}

func (c *cmdPrinter) Infof(format string, i ...interface{}) {
	c.cmd.Printf(format, i...)
}

func (c *cmdPrinter) Warn(a ...interface{}) {
	c.cmd.PrintErr(a...)
}

func (c *cmdPrinter) Warnf(format string, i ...interface{}) {
	c.cmd.PrintErrf(format, i...)
}

type shellPrinter struct {
}

func newShellPrinter() *shellPrinter {
	return &shellPrinter{}
}

func (s *shellPrinter) Info(a ...interface{}) {
	util.Info(a...)
}

func (s *shellPrinter) Infof(format string, i ...interface{}) {
	util.Info(fmt.Sprintf(format, i...))
}

func (s *shellPrinter) Warn(a ...interface{}) {
	util.Warn(a...)
}

func (s *shellPrinter) Warnf(format string, i ...interface{}) {
	util.Warn(fmt.Sprintf(format, i...))
}
