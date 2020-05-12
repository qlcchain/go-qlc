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

func addDSTerminateOrderCmdByShell(parentCmd *ishell.Cmd) {
	buyerAddress := util.Flag{
		Name:  "buyerAddress",
		Must:  true,
		Usage: "buyer's address hex string",
		Value: "",
	}
	buyerName := util.Flag{
		Name:  "buyerName",
		Must:  true,
		Usage: "buyer's name",
		Value: "",
	}
	sellerAddress := util.Flag{
		Name:  "sellerAddress",
		Must:  true,
		Usage: "seller's address",
		Value: "",
	}
	sellerName := util.Flag{
		Name:  "sellerName",
		Must:  true,
		Usage: "seller's name",
		Value: "",
	}
	productId := util.Flag{
		Name:  "productId",
		Must:  true,
		Usage: "productId (separate by comma)",
		Value: "",
	}

	args := []util.Flag{buyerAddress, buyerName, sellerAddress, sellerName, productId}
	cmd := &ishell.Cmd{
		Name:                "terminateOrder",
		Help:                "create a terminate order request",
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

			buyerAddressP := util.StringVar(c.Args, buyerAddress)
			buyerNameP := util.StringVar(c.Args, buyerName)
			sellerAddressP := util.StringVar(c.Args, sellerAddress)
			sellerNameP := util.StringVar(c.Args, sellerName)
			productIdP := util.StringVar(c.Args, productId)

			if err := DSTerminateOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, productIdP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSTerminateOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, productIdP string) error {
	client, err := rpc.Dial(endpointP)
	if err != nil {
		return err
	}
	defer client.Close()

	accBytes, err := hex.DecodeString(buyerAddressP)
	if err != nil {
		return err
	}

	acc := types.NewAccount(accBytes)
	if acc == nil {
		return fmt.Errorf("account format err")
	}

	sellerAddress, err := types.HexToAddress(sellerAddressP)
	if err != nil {
		return err
	}

	param := &abi.DoDSettleTerminateOrderParam{
		Buyer: &abi.DoDSettleUser{
			Address: acc.Address(),
			Name:    buyerNameP,
		},
		Seller: &abi.DoDSettleUser{
			Address: sellerAddress,
			Name:    sellerNameP,
		},
		ProductId: strings.Split(productIdP, ","),
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getTerminateOrderBlock", param)
	if err != nil {
		return err
	}

	var w types.Work
	worker, _ := types.NewWorker(w, block.Root())
	block.Work = worker.NewWork()

	hash := block.GetHash()
	block.Signature = acc.Sign(hash)

	fmt.Printf("block:\n%s\nhash[%s]\ninternalId[%s]\n", cutil.ToIndentString(block), block.GetHash(), block.Previous)

	var h types.Hash
	err = client.Call(&h, "ledger_process", &block)
	if err != nil {
		return err
	}

	return nil
}
