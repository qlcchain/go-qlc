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
	"github.com/qlcchain/go-qlc/config"
	"github.com/spf13/cobra"
)

// wlCmd represents the wl command
var wlCmd = &cobra.Command{
	Use:   "wl",
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func walletList() ([]types.Address, error) {
	if cfgPath == "" {
		cfgPath = config.DefaultConfigFile()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return []types.Address{}, err
	}
	err = initNode(types.ZeroAddress, "", cfg)
	if err != nil {
		return []types.Address{}, err
	}
	w := ctx.Wallet.Wallet

	addresses, err := w.WalletIds()
	if err != nil {
		return []types.Address{}, err
	}
	return addresses, nil
}
