/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

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
	Use:   "walletimport",
	Short: "import a wallet",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := importWallet()
		if err != nil {
			cmd.Println(err)
		} else {
			cmd.Printf("import seed[%s] password[%s] => %s success", seed, pwd, addr.String())
			cmd.Println()
		}
	},
}

func init() {
	wiCmd.Flags().StringVarP(&seed, "seed", "s", "", "seed for a wallet")
	rootCmd.AddCommand(wiCmd)
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
