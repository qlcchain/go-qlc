/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

func blockCount() {
	if interactive {
		c := &ishell.Cmd{
			Name: "blockcount",
			Help: "return the total count of block in db",
			Func: func(c *ishell.Context) {
				if util.HelpText(c, nil) {
					return
				}
				if err := util.CheckArgs(c, nil); err != nil {
					util.Warn(err)
					return
				}
				err := blocks()
				if err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(c)
	} else {
		var blockcountCmd = &cobra.Command{
			Use:   "blockcount",
			Short: "block count",
			Run: func(cmd *cobra.Command, args []string) {
				err := blocks()
				if err != nil {
					cmd.Println(err)
					return
				}

			},
		}
		rootCmd.AddCommand(blockcountCmd)
	}
}

func blocks() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var resp map[string]uint64
	err = client.Call(&resp, "ledger_transactionsCount")
	if err != nil {
		return err
	}

	state := resp["count"]
	unchecked := resp["unchecked"]
	s := fmt.Sprintf("total state block count is: %d, unchecked block count is: %d", state, unchecked)
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}

	return nil

}
