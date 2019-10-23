package commands

import (
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addTxBlockInfoCmdByShell(parentCmd *ishell.Cmd) {
	hashFlag := util.Flag{
		Name:  "hashes",
		Must:  true,
		Usage: "hashes of account blocks",
		Value: "",
	}

	cmd := &ishell.Cmd{
		Name: "getBlockInfo",
		Help: "get account blocks info by hashes",
		Func: func(c *ishell.Context) {
			args := []util.Flag{hashFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			hashStrList := util.StringSliceVar(c.Args, hashFlag)

			err := runTxBlockInfoCmd(hashStrList)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runTxBlockInfoCmd(hashStrList []string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	//rspInfo := make([]*api.APIBlock, len(hashStrList))
	var rspInfo []*api.APIBlock
	err = client.Call(&rspInfo, "ledger_blocksInfo", hashStrList)
	if err != nil {
		return err
	}

	fmt.Println(cutil.ToIndentString(rspInfo))

	return nil
}

func addTxBlockListCmdByShell(parentCmd *ishell.Cmd) {
	addressFlag := util.Flag{
		Name:  "address",
		Must:  false,
		Usage: "address of account hex string",
		Value: "",
	}
	offsetFlag := util.Flag{
		Name:  "offset",
		Must:  false,
		Usage: "start point of total block list",
		Value: 0,
	}
	limitFlag := util.Flag{
		Name:  "limit",
		Must:  false,
		Usage: "count of blocks",
		Value: 10,
	}
	showFlag := util.Flag{
		Name:  "show",
		Must:  false,
		Usage: "show style, etc list, detail",
		Value: "list",
	}

	cmd := &ishell.Cmd{
		Name: "getBlockList",
		Help: "get account blocks list",
		Func: func(c *ishell.Context) {
			args := []util.Flag{addressFlag, offsetFlag, limitFlag, showFlag}
			if util.HelpText(c, args) {
				return
			}
			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			address := util.StringVar(c.Args, addressFlag)
			offset, _ := util.IntVar(c.Args, offsetFlag)
			limit, _ := util.IntVar(c.Args, limitFlag)
			show := util.StringVar(c.Args, showFlag)

			err := runTxBlockListCmd(address, offset, limit, show)
			if err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func runTxBlockListCmd(address string, offset, limit int, show string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	var rspInfo []*api.APIBlock

	if len(address) > 0 {
		err = client.Call(&rspInfo, "ledger_accountHistoryTopn", address, limit, offset)
	} else {
		err = client.Call(&rspInfo, "ledger_blocks", limit, offset)
	}
	if err != nil {
		return err
	}

	if show == "list" {
		fmt.Printf("%-64s %-15s %-10s %-10s %-10s %-10s %s\n",
			"Hash", "Type", "TokenName", "Amount", "PovH", "PovC", "Time")
		for _, apiBlk := range rspInfo {
			fmt.Printf("%-64s %-15s %-10s %-10s %-10d %-10d %s\n",
				apiBlk.Hash, apiBlk.Type,
				apiBlk.TokenName, txFormatBalance(apiBlk.Amount),
				apiBlk.PovConfirmHeight, apiBlk.PovConfirmCount,
				time.Unix(apiBlk.Timestamp, 0).Format("2006-01-02 15:04:05"))
		}
		return nil
	}

	for _, apiBlk := range rspInfo {
		fmt.Printf("\n\n")
		fmt.Printf("%-16s: %s\n", "Hash", apiBlk.Hash)
		fmt.Printf("%-16s: %s\n", "Type", apiBlk.Type)
		fmt.Printf("%-16s: %s\n", "Address", apiBlk.Address)
		if apiBlk.IsSendBlock() {
			toAddr, _ := types.BytesToAddress(apiBlk.Link.Bytes())
			fmt.Printf("%-16s: %s\n", "Link", toAddr)
		} else {
			fmt.Printf("%-16s: %s\n", "Link", apiBlk.Link)
		}
		fmt.Printf("%-16s: %s\n", "Token", apiBlk.TokenName)
		fmt.Printf("%-16s: %s\n", "Amount", apiBlk.Amount)
		fmt.Printf("%-16s: %d(%s)\n", "Time", apiBlk.Timestamp, time.Unix(apiBlk.Timestamp, 0))
		fmt.Printf("%-16s: %d\n", "PovConfirmHeight", apiBlk.PovConfirmHeight)
		fmt.Printf("%-16s: %d\n", "PovConfirmCount", apiBlk.PovConfirmCount)
	}

	return nil
}