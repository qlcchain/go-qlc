package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/rpc/api"
)

func addPermissionNodeRemoveCmdByShell(parentCmd *ishell.Cmd) {
	admin := util.Flag{
		Name:  "admin",
		Must:  true,
		Usage: "admin user (private key in hex string)",
		Value: "",
	}
	index := util.Flag{
		Name:  "index",
		Must:  true,
		Usage: "node index",
		Value: "",
	}
	c := &ishell.Cmd{
		Name: "nodeRemove",
		Help: "permission remove node",
		Func: func(c *ishell.Context) {
			args := []util.Flag{admin, index}
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			adminP := util.StringVar(c.Args, admin)
			indexP, err := util.IntVar(c.Args, index)
			if err != nil {
				util.Warn("parse node index err")
			}

			err = nodeRemove(adminP, indexP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func nodeRemove(adminP string, indexP int) error {
	if adminP == "" {
		return fmt.Errorf("admin can not be null")
	}

	accBytes, err := hex.DecodeString(adminP)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	param := &api.NodeParam{
		Admin: acc.Address(),
		Index: uint32(indexP),
	}

	var block types.StateBlock
	err = client.Call(&block, "permission_getNodeRemoveBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("node remove block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}
