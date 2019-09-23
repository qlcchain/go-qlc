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
	"github.com/qlcchain/go-qlc/cmd/util"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"
)

func dumpledger() {
	if interactive {
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
		shell.AddCmd(c)
	} else {
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
		rootCmd.AddCommand(dumpCmd)
	}
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
	err = client.Call(&path, "debug_dump")
	if err != nil {
		return err
	}

	s = fmt.Sprintf("dump to %s  ", path)
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}

	return nil
}
