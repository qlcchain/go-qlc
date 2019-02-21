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
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

func blockCount() {
	if interactive {
		c := &ishell.Cmd{
			Name: "blockcount",
			Help: "return the total count of block in db",
			Func: func(c *ishell.Context) {
				if HelpText(c, nil) {
					return
				}
				if err := CheckArgs(c, nil); err != nil {
					Warn(err)
					return
				}
				resp, err := blocks()
				if err != nil {
					Warn(err)
					return
				}
				state := resp["count"]
				unchecked := resp["unchecked"]
				Info(fmt.Sprintf("total state block count is: %d, unchecked block count is: %d", state, unchecked))
			},
		}
		shell.AddCmd(c)
	} else {
		var blockcountCmd = &cobra.Command{
			Use:   "blockcount",
			Short: "block count",
			Run: func(cmd *cobra.Command, args []string) {
				resp, err := blocks()
				if err != nil {
					cmd.Println(err)
					return
				}
				state := resp["count"]
				unchecked := resp["unchecked"]
				cmd.Printf("total state block count is: %d, unchecked block count is: %d", state, unchecked)
				cmd.Println()
			},
		}
		rootCmd.AddCommand(blockcountCmd)
	}
}

func blocks() (map[string]uint64, error) {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var resp map[string]uint64
	err = client.Call(&resp, "ledger_transactionsCount")
	if err != nil {
		return nil, err
	}
	return resp, nil
}
