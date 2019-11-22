/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/chain"
	"github.com/qlcchain/go-qlc/chain/context"
	cmdutil "github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common"
)

func purgePov() {
	if interactive {
		startHeight := cmdutil.Flag{
			Name:  "startHeight",
			Must:  true,
			Usage: "pov block start height",
			Value: -1,
		}

		cmd := &ishell.Cmd{
			Name: "purgepov",
			Help: "purge pov blocks from database",
			Func: func(c *ishell.Context) {
				args := []cmdutil.Flag{startHeight}
				if cmdutil.HelpText(c, args) {
					return
				}
				if err := cmdutil.CheckArgs(c, args); err != nil {
					cmdutil.Warn(err)
					return
				}
				startHeightP, _ := cmdutil.IntVar(c.Args, startHeight)
				purgePovAction(startHeightP)
			},
		}
		shell.AddCmd(cmd)
	} else {
		var startHeightP int
		var cmd = &cobra.Command{
			Use:   "purgepov",
			Short: "purge pov blocks from database",
			Run: func(cmd *cobra.Command, args []string) {
				purgePovAction(startHeightP)
			},
		}
		cmd.Flags().IntVarP(&startHeightP, "startHeight", "", -1, "pov block start height")
		rootCmd.AddCommand(cmd)
	}
}

func purgePovAction(startHeight int) {
	if startHeight < 0 {
		fmt.Println("invalid startHeight value")
		return
	}

	var err error

	chainContext := context.NewChainContext(cfgPathP)
	cm, err := chainContext.ConfigManager()
	if err != nil {
		cmdutil.Warn(err)
		return
	}

	cfg, err := cm.Config()
	if err != nil {
		cmdutil.Warn(err)
		return
	}

	cmdutil.Info("ConfigFile", cm.ConfigFile)
	cmdutil.Info("DataDir", cfg.DataDir)

	cmdutil.Info("starting to purge pov blocks, please wait...")
	ledgerService := chain.NewLedgerService(cm.ConfigFile)
	l := ledgerService.Ledger

	latestBlk, err := l.GetLatestPovBlock()
	if err != nil {
		cmdutil.Warn("GetLatestPovBlock", err)
		return
	}

	cmdutil.Info("startHeight", startHeight, "latestHeight", latestBlk.GetHeight())
	if uint64(startHeight) > latestBlk.GetHeight() {
		return
	}

	txlSCH, err := l.GetPovTxlScanCursor()
	if err != nil {
		cmdutil.Warn("GetPovTxlScanCursor", err)
		return
	}

	totalDelBlkCnt := 0
	curHeight := latestBlk.GetHeight()
	for ; curHeight >= uint64(startHeight); curHeight-- {
		povBlk, err := l.GetPovBlockByHeight(curHeight)
		if err != nil {
			cmdutil.Warn("GetPovBlockByHeight", err)
			return
		}

		for _, txPov := range povBlk.GetAllTxs() {
			err = l.DeletePovTxLookup(txPov.GetHash())
			if err != nil {
				cmdutil.Warn("DeletePovTxLookup", err)
			}
		}

		err = l.DeletePovBlock(povBlk)
		if err != nil {
			cmdutil.Warn("DeletePovBlock", err)
		}

		err = l.DeletePovBestHash(curHeight)
		if err != nil {
			cmdutil.Warn("DeletePovBestHash", err)
		}

		totalDelBlkCnt++
		if totalDelBlkCnt%100 == 0 {
			cmdutil.Info("delete pov blocks", "total", totalDelBlkCnt, "curHeight", curHeight)
		}
	}
	if totalDelBlkCnt%100 != 0 {
		cmdutil.Info("delete pov blocks", "total", totalDelBlkCnt, "curHeight", curHeight)
	}

	newLH := uint64(0)
	if startHeight > 0 {
		newLH = uint64(startHeight) - 1
	}

	cmdutil.Info("set pov latest height", newLH)
	err = l.SetPovLatestHeight(newLH)
	if err != nil {
		cmdutil.Warn("SetPovLatestHeight", err)
	}

	if txlSCH > newLH {
		cmdutil.Info("set pov txl scan cursor", newLH)
		err = l.SetPovTxlScanCursor(newLH)
		if err != nil {
			cmdutil.Warn("SetPovTxlScanCursor", err)
		}
	}

	startDay := uint32(startHeight / common.POVChainBlocksPerDay)
	endDay := uint32(latestBlk.GetHeight() / uint64(common.POVChainBlocksPerDay))
	for curDay := startDay; curDay <= endDay; curDay++ {
		cmdutil.Info("delete pov miner stat", "day", curDay)
		err = l.DeletePovMinerStat(curDay)
		if err != nil {
			cmdutil.Warn("DeletePovMinerStat", err)
		}
	}

	cmdutil.Info("finished to purge pov blocks.")
}
