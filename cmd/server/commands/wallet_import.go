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
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/wallet"
)

func addImportWalletCmdByShell(parentCmd *ishell.Cmd) {
	seed := util.Flag{
		Name:  "seed",
		Must:  true,
		Usage: "seed for a wallet",
		Value: "",
	}
	s := &ishell.Cmd{
		Name: "import",
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
	parentCmd.AddCmd(s)
}

func addImportWalletCmdByCobra(parentCmd *cobra.Command) {
	wiCmd := &cobra.Command{
		Use:   "import",
		Short: "import a wallet",
		Run: func(cmd *cobra.Command, args []string) {
			err := importWallet(seedP)
			if err != nil {
				cmd.PrintErr(err)
				return
			}
		},
	}
	wiCmd.Flags().StringVarP(&seedP, "seed", "s", "", "seed for a wallet")
	parentCmd.AddCommand(wiCmd)
}

func importWallet(seedP string) error {
	chain := context.NewChainContext(cfgPathP)
	defer func() {
		if chain != nil {
			_ = chain.Destroy()
			ledger.CloseLedger()
		}
	}()
	cm, err := chain.ConfigManager()
	if err != nil {
		return err
	}
	if len(seedP) != types.SeedSize {
		return errors.New("invalid seed")
	}
	w := wallet.NewWalletStore(cm.ConfigFile)
	defer func() {
		if w != nil {
			_ = w.Close()
		}
	}()

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
