package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSCreateResponseCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "address hex string",
		Value: "",
	}
	hash := util.Flag{
		Name:  "hash",
		Must:  true,
		Usage: "request hash",
		Value: "",
	}
	action := util.Flag{
		Name:  "action",
		Must:  true,
		Usage: "response action (confirm/reject)",
		Value: "",
	}

	args := []util.Flag{address, hash, action}
	cmd := &ishell.Cmd{
		Name:                "createResponse",
		Help:                "response a request",
		CompleterWithPrefix: util.OptsCompleter(args),
		Func: func(c *ishell.Context) {
			if util.HelpText(c, args) {
				return
			}
			err := util.CheckArgs(c, args)
			if err != nil {
				util.Warn(err)
				return
			}

			addressP := util.StringVar(c.Args, address)
			hashP := util.StringVar(c.Args, hash)
			actionP := util.StringVar(c.Args, action)

			if err := DSCreateResponse(addressP, hashP, actionP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSCreateResponse(addressP, hashP, actionP string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	accBytes, err := hex.DecodeString(addressP)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	requestHash, err := types.NewHash(hashP)
	if err != nil {
		return err
	}

	action, err := abi.ParseDoDResponseAction(actionP)
	if err != nil {
		return err
	}

	param := &abi.DoDSettleResponseParam{
		RequestHash: requestHash,
		Action:      action,
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getCreateOrderRewardBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("block:\n%s\nhash[%s]\n", cutil.ToIndentString(block), block.GetHash())

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}
