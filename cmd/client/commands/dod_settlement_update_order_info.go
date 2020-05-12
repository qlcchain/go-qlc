package commands

import (
	"encoding/hex"
	"fmt"
	"math/rand"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSUpdateOrderInfoCmdByShell(parentCmd *ishell.Cmd) {
	buyer := util.Flag{
		Name:  "buyer",
		Must:  true,
		Usage: "buyer's address hex string",
		Value: "",
	}
	internalId := util.Flag{
		Name:  "internalId",
		Must:  true,
		Usage: "order's internalId",
		Value: "",
	}
	orderId := util.Flag{
		Name:  "orderId",
		Must:  true,
		Usage: "order id from sonata api",
		Value: "",
	}
	operation := util.Flag{
		Name:  "operation",
		Must:  true,
		Usage: "sonata operation has been done (create/change/terminate/fail)",
		Value: "",
	}
	reason := util.Flag{
		Name:  "reason",
		Must:  false,
		Usage: "reason of fail",
		Value: "",
	}
	productNum := util.Flag{
		Name:  "productNum",
		Must:  true,
		Usage: "product number",
		Value: "",
	}

	args := []util.Flag{buyer, internalId, orderId, operation, reason, productNum}
	cmd := &ishell.Cmd{
		Name:                "updateOrderInfo",
		Help:                "update order info",
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

			buyerP := util.StringVar(c.Args, buyer)
			internalIdP := util.StringVar(c.Args, internalId)
			orderIdP := util.StringVar(c.Args, orderId)
			operationP := util.StringVar(c.Args, operation)
			reasonP := util.StringVar(c.Args, reason)
			productNumP, err := util.IntVar(c.Args, productNum)
			if err != nil {
				util.Warn(err)
				return
			}

			if err := DSUpdateOrderInfo(buyerP, internalIdP, orderIdP, operationP, reasonP, productNumP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSUpdateOrderInfo(buyerP, internalIdP, orderIdP, operationP, reasonP string, productNum int) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	accBytes, err := hex.DecodeString(buyerP)
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

	operation, err := abi.ParseDoDOrderOperation(operationP)
	if err != nil {
		return err
	}

	param := &abi.DoDSettleUpdateOrderInfoParam{
		Buyer:      acc.Address(),
		InternalId: internalId,
		OrderId:    orderIdP,
		ProductId:  make([]string, 0),
		Operation:  operation,
		FailReason: reasonP,
	}

	for i := 0; i < productNum; i++ {
		param.ProductId = append(param.ProductId, fmt.Sprintf("product%d", rand.Int()))
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getUpdateOrderInfoBlock", param)
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
