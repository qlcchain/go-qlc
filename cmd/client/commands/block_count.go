/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

// bcCmd represents the bc command
var blockcountCmd = &cobra.Command{
	Use:   "blockcount",
	Short: "block count",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := rpc.Dial(endpoint)
		if err != nil {
			cmd.Println(err)
			return
		}
		defer client.Close()

		var resp map[string]uint64
		err = client.Call(&resp, "ledger_transactionsCount")
		if err != nil {
			cmd.Println(err)
			return
		}
		state := resp["count"]
		unchecked := resp["unchecked"]

		cmd.Printf("total stateblock count is : %d, unchecked count is: %d", state, unchecked)
		cmd.Println()
	},
}

func init() {
	rootCmd.AddCommand(blockcountCmd)
}
