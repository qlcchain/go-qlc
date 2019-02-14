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
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/rpc"
	"github.com/spf13/cobra"
)

// wcCmd represents the wc command
var wcCmd = &cobra.Command{
	Use:   "walletcreate",
	Short: "create a wallet for QLCChain node",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := createWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("create wallet: address=>%s, password=>%s success", addr.String(), pwd)
			cmd.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(wcCmd)
}

func createWallet() (types.Address, error) {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return types.ZeroAddress, err
	}
	defer client.Close()
	var addr types.Address
	err = client.Call(&addr, "wallet_newWallet", pwd)
	if err != nil {
		return types.ZeroAddress, err
	}
	return addr, nil
}
