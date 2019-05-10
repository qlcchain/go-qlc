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

	"github.com/qlcchain/go-qlc/cmd/util"

	"github.com/abiosoft/ishell"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/spf13/cobra"
)

func walletimport() {
	var seedP string
	if interactive {
		seed := util.Flag{
			Name:  "seed",
			Must:  true,
			Usage: "seed for a wallet",
			Value: "",
		}
		s := &ishell.Cmd{
			Name: "walletimport",
			Help: "import a wallet",
			Func: func(c *ishell.Context) {
				args := []util.Flag{seed, password, cfgPath}
				if util.HelpText(c, args) {
					return
				}
				if err := util.CheckArgs(c, args); err != nil {
					util.Warn(err)
					return
				}
				seedP = util.StringVar(c.Args, seed)
				passwordP = util.StringVar(c.Args, password)
				cfgPathP = util.StringVar(c.Args, cfgPath)
				//if passwordP = ""
				err := importWallet(seedP)
				if err != nil {
					util.Warn(err)
					return
				}
			},
		}
		shell.AddCmd(s)
	} else {
		var wiCmd = &cobra.Command{
			Use:   "walletimport",
			Short: "import a wallet",
			Run: func(cmd *cobra.Command, args []string) {
				err := importWallet(seedP)
				if err != nil {
					cmd.Println(err)
					return
				}
			},
		}
		wiCmd.Flags().StringVarP(&seedP, "seed", "s", "", "seed for a wallet")
		rootCmd.AddCommand(wiCmd)
	}
}

func importWallet(seedP string) error {
	cfg, err := cmdutil.GetConfig(cfgPathP)
	if err != nil {
		return err
	}
	if len(seedP) == 0 {
		return errors.New("invalid seed")
	}
	var accounts []*types.Account
	err = initNode(accounts, cfg)
	if err != nil {
		return err
	}
	w := walletService.Wallet
	addr, err := w.NewWalletBySeed(seedP, passwordP)
	if err != nil {
		return err
	}

	s := fmt.Sprintf("import seed[%s] password[%s] => %s success", seedP, passwordP, addr.String())
	if interactive {
		util.Info(s)
	} else {
		fmt.Println(s)
	}
	return nil
}
