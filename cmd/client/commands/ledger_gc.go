/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"github.com/qlcchain/go-qlc/common/storage"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
)

func addLedgerGCByIshell(parentCmd *ishell.Cmd) {
	c := &ishell.Cmd{
		Name: "gc",
		Help: "badger value log garbage collection",
		Func: func(c *ishell.Context) {
			if util.HelpText(c, nil) {
				return
			}
			if err := util.CheckArgs(c, nil); err != nil {
				util.Warn(err)
				return
			}
			if err := gc(); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(c)
}

func addLedgerGCByCobra(parentCmd *cobra.Command) {
	var dumpCmd = &cobra.Command{
		Use:   "gc",
		Short: "badger value log garbage collection",
		Run: func(cmd *cobra.Command, args []string) {
			err := gc()
			if err != nil {
				cmd.Println(err)
				return
			}
		},
	}
	parentCmd.AddCommand(dumpCmd)
}

func gc() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var path string
	err = client.Call(&path, "debug_action", storage.GC, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	s := fmt.Sprintf("ledger gc done ")
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}

	return nil
}
