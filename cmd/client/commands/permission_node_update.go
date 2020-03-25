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

func addPermissionNodeUpdateCmdByShell(parentCmd *ishell.Cmd) {
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
	kind := util.Flag{
		Name:  "kind",
		Must:  true,
		Usage: "node kind (0:ip 1:peer id)",
		Value: "",
	}
	node := util.Flag{
		Name:  "node",
		Must:  true,
		Usage: "node addr",
		Value: "",
	}
	comment := util.Flag{
		Name:  "comment",
		Must:  false,
		Usage: "node comment",
		Value: "",
	}
	args := []util.Flag{admin, index, kind, node, comment}
	c := &ishell.Cmd{
		Name:                "nodeUpdate",
		Help:                "permission update node",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}

			if err := util.CheckArgs(c, args); err != nil {
				util.Warn(err)
				return
			}

			adminP := util.StringVar(c.Args, admin)
			nodeP := util.StringVar(c.Args, node)
			commentP := util.StringVar(c.Args, comment)
			kindP, err := util.IntVar(c.Args, kind)
			if err != nil {
				util.Warn("parse node kind err")
			}
			indexP, err := util.IntVar(c.Args, index)
			if err != nil {
				util.Warn("parse node index err")
			}

			err = nodeUpdate(adminP, indexP, kindP, nodeP, commentP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func nodeUpdate(adminP string, indexP int, kindP int, nodeP string, commentP string) error {
	if adminP == "" {
		return fmt.Errorf("admin can not be null")
	}

	if nodeP == "" {
		return fmt.Errorf("node addr can not be null")
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
		Admin:   acc.Address(),
		Index:   uint32(indexP),
		Kind:    uint8(kindP),
		Node:    nodeP,
		Comment: commentP,
	}

	var block types.StateBlock
	err = client.Call(&block, "permission_getNodeUpdateBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("node update block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}
