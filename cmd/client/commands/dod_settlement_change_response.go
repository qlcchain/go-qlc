package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/qlcchain/go-qlc/rpc/api"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSChangeResponseCmdByShell(parentCmd *ishell.Cmd) {
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
	privateFrom := util.Flag{
		Name:  "privateFrom",
		Must:  false,
		Usage: "privateFrom",
		Value: "",
	}
	privateFor := util.Flag{
		Name:  "privateFor",
		Must:  false,
		Usage: "privateFor",
		Value: "",
	}

	args := []util.Flag{address, hash, action, privateFrom, privateFor}
	cmd := &ishell.Cmd{
		Name:                "changeResponse",
		Help:                "response change request",
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
			privateFromP := util.StringVar(c.Args, privateFrom)
			privateForP := util.StringVar(c.Args, privateFor)

			if err := DSChangeResponse(addressP, hashP, actionP, privateFromP, privateForP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSChangeResponse(addressP, hashP, actionP, privateFromP, privateForP string) error {
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

	action, err := abi.ParseDoDSettleResponseAction(actionP)
	if err != nil {
		return err
	}

	param := &api.DoDSettleResponseParam{
		ContractPrivacyParam: api.ContractPrivacyParam{
			PrivateFrom: privateFromP,
			PrivateFor:  strings.Split(privateForP, ","),
		},
		DoDSettleResponseParam: abi.DoDSettleResponseParam{
			RequestHash: requestHash,
			Action:      action,
		},
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getChangeOrderRewardBlock", param)
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
