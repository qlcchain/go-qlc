/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/cmd/client/commands"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/spf13/cobra"
)

func walletimport() {
	var seedP string
	if interactive {
		seed := commands.Flag{
			Name:  "seed",
			Must:  true,
			Usage: "seed for a wallet",
			Value: "",
		}
		s := &ishell.Cmd{
			Name: "walletimport",
			Help: "import a wallet",
			Func: func(c *ishell.Context) {
				args := []commands.Flag{seed, password, cfgPath}
				if commands.HelpText(c, args) {
					return
				}
				if err := commands.CheckArgs(c, args); err != nil {
					commands.Warn(err)
					return
				}
				seedP = commands.StringVar(c.Args, seed)
				passwordP = commands.StringVar(c.Args, password)
				cfgPathP = commands.StringVar(c.Args, cfgPath)
				//if passwordP = ""
				addr, err := importWallet(seedP)
				if err != nil {
					commands.Warn(err)
				} else {
					commands.Info(fmt.Sprintf("import seed[%s] password[%s] => %s success", seedP, passwordP, addr.String()))
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var wiCmd = &cobra.Command{
			Use:   "walletimport",
			Short: "import a wallet",
			Run: func(cmd *cobra.Command, args []string) {
				addr, err := importWallet(seedP)
				if err != nil {
					cmd.Println(err)
				} else {
					cmd.Printf("import seed[%s] password[%s] => %s success", seedP, passwordP, addr.String())
					cmd.Println()
				}
			},
		}
		wiCmd.Flags().StringVarP(&seedP, "seed", "s", "", "seed for a wallet")
		rootCmd.AddCommand(wiCmd)
	}
}

func importWallet(seedP string) (types.Address, error) {
	if len(seedP) == 0 {
		return types.ZeroAddress, errors.New("invalid seed")
	}
	if cfgPathP == "" {
		cfgPathP = config.DefaultDataDir()
	}
	cm := config.NewCfgManager(cfgPathP)
	cfg, err := cm.Load()
	if err != nil {
		return types.ZeroAddress, err
	}
	err = initNode(types.ZeroAddress, "", cfg)
	if err != nil {
		return types.ZeroAddress, err
	}
	w := ctx.Wallet.Wallet
	if addr, err := w.NewWalletBySeed(seedP, passwordP); err != nil {
		return types.ZeroAddress, err
	} else {
		return addr, nil
	}
}
