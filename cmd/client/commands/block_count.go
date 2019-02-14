// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		}
		defer client.Close()

		var resp map[string]uint64
		err = client.Call(&resp, "ledger_transactionsCount")
		if err != nil {
			cmd.Println(err)
		}
		state := resp["count"]
		unchecked := resp["unchecked"]

		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("total stateblock count is : %d, unchecked count is: %d", state, unchecked)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(blockcountCmd)
}
