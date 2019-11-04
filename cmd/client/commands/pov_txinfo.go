package commands

import (
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPovTxInfoCmdByShell(parentCmd *ishell.Cmd) {
	txHashFlag := util.Flag{
		Name:  "txHash",
		Must:  true,
		Usage: "hash of transaction",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getTxInfo",
		Help: "get transaction info",
		Func: func(c *ishell.Context) {
			args := []util.Flag{txHashFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			txHashStr := util.StringVar(c.Args, txHashFlag)

			err := runPovTxInfoCmd(txHashStr)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runPovTxInfoCmd(txHashStr string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	rspInfo := new(api.PovApiTxLookup)
	err = client.Call(rspInfo, "pov_getTransaction", txHashStr)
	if err != nil {
		return err
	}

	fmt.Println(cutil.ToIndentString(rspInfo))

	/*
		fmt.Printf("TxHash:      %s\n", rspInfo.TxHash)
		fmt.Printf("BlockHash:   %s\n", rspInfo.TxLookup.BlockHash)
		fmt.Printf("BlockHeight: %d\n", rspInfo.TxLookup.BlockHeight)
		fmt.Printf("TxIndex:     %d\n", rspInfo.TxLookup.TxIndex)

		if rspInfo.CoinbaseTx != nil {
		} else if rspInfo.AccountTx != nil {
		}
	*/

	return nil
}
