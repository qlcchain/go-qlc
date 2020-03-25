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

func addPermissionNodeAddCmdByShell(parentCmd *ishell.Cmd) {
	admin := util.Flag{
		Name:  "admin",
		Must:  true,
		Usage: "admin user (private key in hex string)",
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
	args := []util.Flag{admin, kind, node, comment}
	c := &ishell.Cmd{
		Name:                "nodeAdd",
		Help:                "permission add node",
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

			err = nodeAdd(adminP, kindP, nodeP, commentP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func nodeAdd(adminP string, kindP int, nodeP string, commentP string) error {
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
		Kind:    uint8(kindP),
		Node:    nodeP,
		Comment: commentP,
	}

	var block types.StateBlock
	err = client.Call(&block, "permission_getNodeAddBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("node add block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}
