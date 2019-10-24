package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"
	"github.com/spf13/cobra"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPovMiningInfoCmdByCobra(parentCmd *cobra.Command) {
}

func addPovMiningInfoCmdByShell(parentCmd *ishell.Cmd) {
	cmd := &ishell.Cmd{
		Name: "getMiningInfo",
		Help: "get mining info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{}
			if util.HelpText(c, args) {
				return
			}

			err := runPovMiningInfoCmd()
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovMiningInfoCmd() error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiGetMiningInfo)
	err = client.Call(rspInfo, "pov_getMiningInfo")
	if err != nil {
		return err
	}

	ss := common.SyncState(rspInfo.SyncState)
	fmt.Printf("Init Sync State: %d(%s)\n", rspInfo.SyncState, ss)
	fmt.Printf("Difficulty: %s\n", formatPovDifficulty(rspInfo.Difficulty))
	fmt.Printf("Pending Txs In Pool: %d\n", rspInfo.PooledTx)

	fmt.Printf("Current Block Height: %d\n", rspInfo.CurrentBlockHeight)
	fmt.Printf("Current Block Hash:   %s\n", rspInfo.CurrentBlockHash)
	fmt.Printf("Current Block Tx:     %d\n", rspInfo.CurrentBlockTx)
	fmt.Printf("Current Block Size:   %d\n", rspInfo.CurrentBlockSize)

	fmt.Printf("Network Hashes Per Second In Last 120 Blocks:\n")
	fmt.Printf(" Chain:   %d\n", rspInfo.HashInfo.ChainHashPS)
	fmt.Printf(" SHA256D: %d\n", rspInfo.HashInfo.Sha256dHashPS)
	fmt.Printf(" X11:     %d\n", rspInfo.HashInfo.X11HashPS)
	fmt.Printf(" SCRYPT:  %d\n", rspInfo.HashInfo.ScryptHashPS)

	rspLs := new(api.PovLedgerStats)
	err = client.Call(rspLs, "pov_getLedgerStats")
	if err != nil {
		return err
	}

	fmt.Printf("Ledger Statistics:\n")
	fmt.Printf(" PovBestCount:    %d\n", rspLs.PovBestCount)
	fmt.Printf(" PovBlockCount:   %d\n", rspLs.PovBlockCount)
	fmt.Printf(" PovAllTxCount:   %d\n", rspLs.PovAllTxCount)
	fmt.Printf(" PovCbTxCount:    %d\n", rspLs.PovCbTxCount)
	fmt.Printf(" PovStateTxCount: %d\n", rspLs.PovStateTxCount)
	fmt.Printf(" StateBlockCount: %d\n", rspLs.StateBlockCount)

	return nil
}
