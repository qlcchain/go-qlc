package commands

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"

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
	orderStatus := util.Flag{
		Name:  "orderStatus",
		Must:  true,
		Usage: "sonata api status (success/fail)",
		Value: "",
	}
	reason := util.Flag{
		Name:  "reason",
		Must:  false,
		Usage: "reason of fail",
		Value: "",
	}
	productBuyerIds := util.Flag{
		Name:  "productBuyerIds",
		Must:  false,
		Usage: "buyer product ids (separate by comma)",
		Value: "",
	}
	productIds := util.Flag{
		Name:  "productIds",
		Must:  false,
		Usage: "product ids (separate by comma)",
		Value: "",
	}

	args := []util.Flag{buyer, internalId, orderId, orderStatus, reason, productBuyerIds, productIds}
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
			orderStatusP := util.StringVar(c.Args, orderStatus)
			reasonP := util.StringVar(c.Args, reason)
			productBuyerIdsP := util.StringVar(c.Args, productBuyerIds)
			productIdsP := util.StringVar(c.Args, productIds)

			if err := DSUpdateOrderInfo(buyerP, internalIdP, orderIdP, orderStatusP, reasonP, productBuyerIdsP, productIdsP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSUpdateOrderInfo(buyerP, internalIdP, orderIdP, orderStatusP, reasonP, productBuyerIdsP, productIdsP string) error {
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

	orderStatus, err := abi.ParseDoDSettleOrderState(orderStatusP)
	if err != nil {
		return err
	}

	productBuyerIds := strings.Split(productBuyerIdsP, ",")
	productIds := strings.Split(productIdsP, ",")

	param := &abi.DoDSettleUpdateOrderInfoParam{
		Buyer:      acc.Address(),
		InternalId: internalId,
		OrderId:    orderIdP,
		ProductIds: make([]*abi.DoDSettleProductItem, 0),
		Status:     orderStatus,
		FailReason: reasonP,
	}

	for i := 0; i < len(productBuyerIds); i++ {
		pi := &abi.DoDSettleProductItem{
			ProductId:      productIds[i],
			BuyerProductId: productBuyerIds[i],
			OrderItemId:    fmt.Sprintf("orderItem%d", rand.Int31n(100)),
		}
		param.ProductIds = append(param.ProductIds, pi)
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
