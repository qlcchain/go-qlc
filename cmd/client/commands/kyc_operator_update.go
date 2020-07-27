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

func addKYOperatorUpdateCmdByShell(parentCmd *ishell.Cmd) {
	admin := util.Flag{
		Name:  "admin",
		Must:  true,
		Usage: "admin user (private key in hex string)",
		Value: "",
	}
	operator := util.Flag{
		Name:  "operator",
		Must:  true,
		Usage: "operator address",
		Value: "",
	}
	action := util.Flag{
		Name:  "action",
		Must:  false,
		Usage: "add/remove operator",
		Value: "",
	}
	comment := util.Flag{
		Name:  "comment",
		Must:  false,
		Usage: "operator comment",
		Value: "",
	}
	args := []util.Flag{admin, operator, action, comment}
	c := &ishell.Cmd{
		Name:                "operatorUpdate",
		Help:                "update operator",
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
			operatorP := util.StringVar(c.Args, operator)
			actionP := util.StringVar(c.Args, action)
			commentP := util.StringVar(c.Args, comment)

			err := operatorUpdate(adminP, operatorP, actionP, commentP)
			if err != nil {
				util.Warn(err)
			}
		},
	}
	parentCmd.AddCmd(c)
}

func operatorUpdate(adminP, operatorP, actionP, commentP string) error {
	if adminP == "" {
		return fmt.Errorf("admin can not be null")
	}

	if operatorP == "" {
		return fmt.Errorf("address can not be null")
	}

	accBytes, err := hex.DecodeString(adminP)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	operator, err := types.HexToAddress(operatorP)
	if err != nil {
		return fmt.Errorf("address format err")
	}

	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	param := &api.KYCUpdateOperatorParam{
		Admin:    acc.Address(),
		Operator: operator,
		Action:   actionP,
		Comment:  commentP,
	}

	var block types.StateBlock
	err = client.Call(&block, "KYC_getUpdateOperatorBlock", param)
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
