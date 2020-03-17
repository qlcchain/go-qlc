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
	"github.com/spf13/cobra"

	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
)

func removeDB() {
	if interactive {
		dataType := cmdutil.Flag{
			Name:  "dataType",
			Must:  false,
			Usage: "data type, etc all/pov",
			Value: "all",
		}

		cmdRm := &ishell.Cmd{
			Name: "removedb",
			Help: "remove database",
			Func: func(c *ishell.Context) {
				args := []cmdutil.Flag{dataType}
				if cmdutil.HelpText(c, args) {
					return
				}
				if err := cmdutil.CheckArgs(c, args); err != nil {
					cmdutil.Warn(err)
					return
				}
				dataTypeP := cmdutil.StringVar(c.Args, dataType)
				removeDBAction(dataTypeP)
			},
		}
		shell.AddCmd(cmdRm)
	} else {
		var dataTypeP string
		var cmdRm = &cobra.Command{
			Use:   "removedb",
			Short: "remove database",
			Run: func(cmd *cobra.Command, args []string) {
				removeDBAction(dataTypeP)
			},
		}
		cmdRm.Flags().StringVarP(&dataTypeP, "dataType", "", "all", "data type, etc all/pov")
		rootCmd.AddCommand(cmdRm)
	}
}

func removeDBAction(dataTypeP string) {
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

	if dataTypeP == "all" {
		cmdutil.Info("drop all data in ledger ...")

		if err := ledgerService.Ledger.DBStore().Drop(nil); err != nil {
			cmdutil.Warn(err)
			return
		}
	} else if dataTypeP == "pov" {
		cmdutil.Info("drop all pov in ledger ...")

		if err := ledgerService.Ledger.DropAllPovBlocks(); err != nil {
			cmdutil.Warn(err)
			return
		}
	}

	cmdutil.Info("finished to remove database.")
}
