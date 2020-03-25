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
	"time"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/storage"
)

func addLedgerDumpByIshell(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "dump",
		Help: "dump ledger to sqlite",
		Func: func(c *ishell.Context) {
			if util.HelpText(c, nil) {
				return
			}
			if err := util.CheckArgs(c, nil); err != nil {
				util.Warn(err)
				return
			}
			err := dump()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addLedgerDumpByCobra(parentCmd *cobra.Command) {
	var dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "dump ledger to sqlite",
		Run: func(cmd *cobra.Command, args []string) {
			err := dump()
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	parentCmd.AddCommand(dumpCmd)
}

func dump() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	s := fmt.Sprintf("dump ledger ...")
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}

	var path string
	err = client.Call(&path, "debug_action", storage.Dump, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(fmt.Sprintf("dump to %s  ", path))

	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			var r string
			err = client.Call(&r, "debug_action", storage.Dump, 1)
			if err != nil {
				return err
			}
			if strings.EqualFold(r, "done") {
				s = fmt.Sprintf("dump successfully  ")
				if interactive {
					util.Info(s)
				} else {
					fmt.Println(s)
				}
				return nil
			}
		}
	}
}
