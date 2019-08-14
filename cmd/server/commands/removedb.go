/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"github.com/qlcchain/go-qlc/chain"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/ledger/db"

	"github.com/abiosoft/ishell"
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
	cmdutil.Info("starting to remove database, please wait...")
	ledgerService := chain.NewLedgerService(cfgPathP)
	cmdutil.Info("drop all data in ledger ...")
	err := ledgerService.Ledger.BatchUpdate(func(txn db.StoreTxn) error {
		return txn.Drop(nil)
	})
	if err != nil {
		cmdutil.Warn(err)
		return
	}
	sqliteService, err := chain.NewSqliteService(cfgPathP)
	if sqliteService != nil {
		cmdutil.Info("drop all data in relation ...")
		err = sqliteService.Relation.EmptyStore()
		if err != nil {
			cmdutil.Warn(err)
			return
		}
	}

	cmdutil.Info("finished to remove database.")
}
