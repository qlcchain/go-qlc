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

// wcCmd represents the wc command
var wcCmd = &cobra.Command{
	Use:   "wc",
	Short: "create a wallet for QLCChain node",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := createWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("create wallet: address=>%s, password=>%s success", addr.String(), pwd)
		}
	},
}

func init() {
	rootCmd.AddCommand(wcCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createWallet() (types.Address, error) {
	if cfgPath == "" {
		cfgPath = config.DefaultConfigFile()
	}
	cm := config.NewCfgManager(cfgPath)
	cfg, err := cm.Load()
	if err != nil {
		return types.ZeroAddress, err
	}
	err = initNode(types.ZeroAddress, "", cfg)
	if err != nil {
		return types.ZeroAddress, err
	}
	w := ctx.Wallet.Wallet
	address, err := w.NewWallet()
	if err != nil {
		return types.ZeroAddress, err
	}

	if len(pwd) > 0 {
		if err := w.NewSession(address).ChangePassword(pwd); err != nil {
			return types.ZeroAddress, err
		}
	}
	return address, nil
}
