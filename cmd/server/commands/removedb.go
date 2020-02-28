/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"
	"github.com/qlcchain/go-qlc/chain"

	"github.com/qlcchain/go-qlc/chain/context"

	"github.com/abiosoft/ishell"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/spf13/cobra"
)

func removeDB() {
	if interactive {
		cmdRm := &ishell.Cmd{
			Name: "removedb",
			Help: "remove database",
			Func: func(c *ishell.Context) {
				if cmdutil.HelpText(c, nil) {
					return
				}
				if err := cmdutil.CheckArgs(c, nil); err != nil {
					cmdutil.Warn(err)
					return
				}
				removeDBAction()
			},
		}
		shell.AddCmd(cmdRm)
	} else {
		var cmdRm = &cobra.Command{
			Use:   "removedb",
			Short: "remove database",
			Run: func(cmd *cobra.Command, args []string) {
				removeDBAction()
			},
		}
		rootCmd.AddCommand(cmdRm)
	}
}

func removeDBAction() {
	var err error

	chainContext := context.NewChainContext(cfgPathP)
	cm, err := chainContext.ConfigManager()
	if err != nil {
		fmt.Println(err)
		return
	}

	cfg, err := cm.Config()
	if err != nil {
		fmt.Println(err)
		return
	}

	cmdutil.Info("ConfigFile", cm.ConfigFile)
	cmdutil.Info("DataDir", cfg.DataDir)

	cmdutil.Info("starting to remove database, please wait...")
	ledgerService := chain.NewLedgerService(cm.ConfigFile)
	cmdutil.Info("drop all data in ledger ...")

	if err := ledgerService.Ledger.DBStore().Drop(nil); err != nil {
		cmdutil.Warn(err)
		return
	}

	cmdutil.Info("finished to remove database.")
}
