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
	nodeId := util.Flag{
		Name:  "nodeId",
		Must:  true,
		Usage: "node id",
		Value: "",
	}
	nodeUrl := util.Flag{
		Name:  "nodeUrl",
		Must:  false,
		Usage: "node url",
		Value: "",
	}
	comment := util.Flag{
		Name:  "comment",
		Must:  false,
		Usage: "node comment",
		Value: "",
	}
	args := []util.Flag{admin, nodeId, nodeUrl, comment}
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
			nodeIdP := util.StringVar(c.Args, nodeId)
			nodeUrlP := util.StringVar(c.Args, nodeUrl)
			commentP := util.StringVar(c.Args, comment)

			err := nodeUpdate(adminP, nodeIdP, nodeUrlP, commentP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func nodeUpdate(admin, nodeId, nodeUrl, comment string) error {
	if admin == "" {
		return fmt.Errorf("admin can not be null")
	}

	if nodeId == "" {
		return fmt.Errorf("node id can not be null")
	}

	accBytes, err := hex.DecodeString(admin)
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
		NodeId:  nodeId,
		NodeUrl: nodeUrl,
		Comment: comment,
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
