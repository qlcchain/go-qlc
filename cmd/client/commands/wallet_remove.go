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

// wrCmd represents the wr command
var wrCmd = &cobra.Command{
	Use:   "walletremove",
	Short: "remove wallet",
	Run: func(cmd *cobra.Command, args []string) {
		err := removeWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("remove wallet: %s success", account)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(wrCmd)
}

func removeWallet() error {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return err
	}
	defer client.Close()
	err = client.Call(nil, "wallet_remove", account)
	if err != nil {
		return err
	}
	return nil
}
