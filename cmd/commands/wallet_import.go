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
	"errors"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"

	"github.com/spf13/cobra"
)

var seed string

// wiCmd represents the wi command
var wiCmd = &cobra.Command{
	Use:   "wi",
	Short: "import a wallet",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := importWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("import seed[%s] password[%s] => %s success", seed, pwd, addr.String())
		}
	},
}

func init() {
	wiCmd.Flags().StringVarP(&seed, "seed", "s", "", "seed for a wallet")
	rootCmd.AddCommand(wiCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// wiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func importWallet() (types.Address, error) {
	if len(seed) == 0 {
		return types.ZeroAddress, errors.New("invalid seed")
	}
	if cfgPath == "" {
		cfgPath = config.DefaultDataDir()
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
	if addr, err := w.NewWalletBySeed(seed, pwd); err != nil {
		return types.ZeroAddress, err
	} else {
		return addr, nil
	}
}
