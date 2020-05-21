package commands

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSResourceReadyCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "address hex string",
		Value: "",
	}
	internalId := util.Flag{
		Name:  "internalId",
		Must:  true,
		Usage: "internalId",
		Value: "",
	}
	productId := util.Flag{
		Name:  "productId",
		Must:  true,
		Usage: "productId (separate by comma)",
		Value: "",
	}

	args := []util.Flag{address, internalId, productId}
	cmd := &ishell.Cmd{
		Name:                "resourceReady",
		Help:                "notify resource is ready",
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
			internalIdP := util.StringVar(c.Args, internalId)
			productIdP := util.StringVar(c.Args, productId)

			if err := DSResourceReady(addressP, internalIdP, productIdP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSResourceReady(addressP, internalIdP, productIdP string) error {
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

	internalId, err := types.NewHash(internalIdP)
	if err != nil {
		return err
	}

	param := &abi.DoDSettleResourceReadyParam{
		Address:    acc.Address(),
		InternalId: internalId,
		ProductId:  strings.Split(productIdP, ","),
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getResourceReadyBlock", param)
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
