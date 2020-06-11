package commands

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/qlcchain/go-qlc/rpc/api"

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
	price := util.Flag{
		Name:  "price",
		Must:  true,
		Usage: "price",
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

	args := []util.Flag{buyerAddress, buyerName, sellerAddress, sellerName, productId, price, privateFrom, privateFor}
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
			priceP := util.StringVar(c.Args, price)
			privateFromP := util.StringVar(c.Args, privateFrom)
			privateForP := util.StringVar(c.Args, privateFor)

			if err := DSTerminateOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, productIdP,
				priceP, privateFromP, privateForP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSTerminateOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, productIdP, priceP, privateFromP, privateForP string) error {
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

	price, err := strconv.ParseFloat(priceP, 64)
	if err != nil {
		return err
	}

	param := &api.DoDSettleTerminateOrderParam{
		ContractPrivacyParam: api.ContractPrivacyParam{
			PrivateFrom: privateFromP,
			PrivateFor:  strings.Split(privateForP, ","),
		},
		DoDSettleTerminateOrderParam: abi.DoDSettleTerminateOrderParam{
			Buyer: &abi.DoDSettleUser{
				Address: acc.Address(),
				Name:    buyerNameP,
			},
			Seller: &abi.DoDSettleUser{
				Address: sellerAddress,
				Name:    sellerNameP,
			},
			Connections: make([]*abi.DoDSettleChangeConnectionParam, 0),
		},
	}

	pids := strings.Split(productIdP, ",")

	for _, productId := range pids {
		conn := new(abi.DoDSettleChangeConnectionParam)
		conn.Price = price
		conn.Currency = "USD"
		conn.ProductId = productId
		conn.QuoteId = fmt.Sprintf("quote%d", rand.Int())
		conn.QuoteItemId = fmt.Sprintf("quoteItem%d", rand.Int())
		conn.ItemId = fmt.Sprintf("itemid%d", rand.Int31n(100))
		param.Connections = append(param.Connections, conn)

		fmt.Println(conn.ItemId)
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
