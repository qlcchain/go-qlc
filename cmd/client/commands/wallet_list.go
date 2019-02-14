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

// wlCmd represents the wl command
var wlCmd = &cobra.Command{
	Use:   "walletlist",
	Short: "wallet address list",
	Run: func(cmd *cobra.Command, args []string) {
		addrs, err := walletList()
		if err != nil {
			cmd.Println(err)
		} else {
			if len(addrs) == 0 {
				cmd.Println("no account ,you can try import one!")
			} else {
				for _, v := range addrs {
					cmd.Println(v)
				}
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(wlCmd)
}

func walletList() ([]types.Address, error) {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	var addresses []types.Address
	err = client.Call(&addresses, "wallet_list", pwd)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}
