package commands

import (
	"encoding/hex"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/qlcchain/go-qlc/rpc/api"
	rpc "github.com/qlcchain/jsonrpc2"
	"strconv"
	"strings"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSUpdateProductInfoCmdByShell(parentCmd *ishell.Cmd) {
	address := util.Flag{
		Name:  "address",
		Must:  true,
		Usage: "address hex string",
		Value: "",
	}
	orderId := util.Flag{
		Name:  "orderId",
		Must:  true,
		Usage: "orderId",
		Value: "",
	}
	productId := util.Flag{
		Name:  "productId",
		Must:  true,
		Usage: "productId (separate by comma)",
		Value: "",
	}
	orderItemId := util.Flag{
		Name:  "orderItemId",
		Must:  true,
		Usage: "orderItemId",
		Value: "",
	}
	active := util.Flag{
		Name:  "active",
		Must:  true,
		Usage: "active",
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

	args := []util.Flag{address, orderId, productId, orderItemId, active, privateFrom, privateFor}
	cmd := &ishell.Cmd{
		Name:                "updateProductInfo",
		Help:                "update product info",
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
			orderIdP := util.StringVar(c.Args, orderId)
			productIdP := util.StringVar(c.Args, productId)
			orderItemIdP := util.StringVar(c.Args, orderItemId)
			activeP := util.StringVar(c.Args, active)
			privateFromP := util.StringVar(c.Args, privateFrom)
			privateForP := util.StringVar(c.Args, privateFor)

			if err := DSUpdateProductInfo(addressP, orderIdP, productIdP, orderItemIdP, activeP, privateFromP, privateForP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSUpdateProductInfo(addressP, orderIdP, productIdP, orderItemIdP, activeP, privateFromP, privateForP string) error {
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

	active, err := strconv.ParseBool(activeP)
	if err != nil {
		return err
	}

	param := &api.DoDSettleUpdateProductInfoParam{
		ContractPrivacyParam: api.ContractPrivacyParam{
			PrivateFrom: privateFromP,
			PrivateFor:  strings.Split(privateForP, ","),
		},
		DoDSettleUpdateProductInfoParam: abi.DoDSettleUpdateProductInfoParam{
			Address: acc.Address(),
			OrderId: orderIdP,
			ProductInfo: []*abi.DoDSettleProductInfo{{
				OrderItemId: orderItemIdP,
				ProductId:   productIdP,
				Active:      active,
			}},
		},
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getUpdateProductInfoBlock", param)
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
