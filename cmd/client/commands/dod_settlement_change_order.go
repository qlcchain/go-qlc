package commands

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSChangeOrderCmdByShell(parentCmd *ishell.Cmd) {
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
	billingType := util.Flag{
		Name:  "billingType",
		Must:  true,
		Usage: "billing type (PAYG/DOD)",
		Value: "",
	}
	bandwidth := util.Flag{
		Name:  "bandwidth",
		Must:  true,
		Usage: "connection bandwidth (10 Mbps)",
		Value: "",
	}
	billingUnit := util.Flag{
		Name:  "billingUnit",
		Must:  false,
		Usage: "billing unit (year/month/week/day/hour/minute/second)",
		Value: "",
	}
	price := util.Flag{
		Name:  "price",
		Must:  true,
		Usage: "price",
		Value: "",
	}
	startTime := util.Flag{
		Name:  "startTime",
		Must:  false,
		Usage: "startTime",
		Value: "",
	}
	endTime := util.Flag{
		Name:  "endTime",
		Must:  false,
		Usage: "endTime",
		Value: "",
	}
	productId := util.Flag{
		Name:  "productId",
		Must:  true,
		Usage: "productId (separate by comma)",
		Value: "",
	}

	args := []util.Flag{buyerAddress, buyerName, sellerAddress, sellerName, billingType, bandwidth, billingUnit, price,
		startTime, endTime, productId}
	cmd := &ishell.Cmd{
		Name:                "changeOrder",
		Help:                "create a change order request",
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
			billingTypeP := util.StringVar(c.Args, billingType)
			bandwidthP := util.StringVar(c.Args, bandwidth)
			billingUnitP := util.StringVar(c.Args, billingUnit)
			priceP := util.StringVar(c.Args, price)
			startTimeP := util.StringVar(c.Args, startTime)
			endTimeP := util.StringVar(c.Args, endTime)
			productIdP := util.StringVar(c.Args, productId)

			if err := DSChangeOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, startTimeP, endTimeP,
				billingTypeP, bandwidthP, billingUnitP, priceP, productIdP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSChangeOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, startTimeP, endTimeP, billingTypeP,
	bandwidthP, billingUnitP, priceP, productIdP string) error {
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

	billingType, err := abi.ParseDoDSettleBillingType(billingTypeP)
	if err != nil {
		return err
	}

	price, err := strconv.ParseFloat(priceP, 64)
	if err != nil {
		return err
	}

	var startTime, endTime int64
	var billingUnit abi.DoDSettleBillingUnit

	if billingType == abi.DoDSettleBillingTypeDOD {
		startTime, err = strconv.ParseInt(startTimeP, 10, 64)
		if err != nil {
			return err
		}

		endTime, err = strconv.ParseInt(endTimeP, 10, 64)
		if err != nil {
			return err
		}
	} else {
		billingUnit, err = abi.ParseDoDSettleBillingUnit(billingUnitP)
		if err != nil {
			return err
		}
	}

	param := &abi.DoDSettleChangeOrderParam{
		Buyer: &abi.DoDSettleUser{
			Address: acc.Address(),
			Name:    buyerNameP,
		},
		Seller: &abi.DoDSettleUser{
			Address: sellerAddress,
			Name:    sellerNameP,
		},
		Connections: make([]*abi.DoDSettleChangeConnectionParam, 0),
	}

	pids := strings.Split(productIdP, ",")

	for _, productId := range pids {
		var conn *abi.DoDSettleChangeConnectionParam

		if billingType == abi.DoDSettleBillingTypePAYG {
			conn = &abi.DoDSettleChangeConnectionParam{
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					ItemId:      fmt.Sprintf("itemid%d", rand.Int31n(100)),
					Bandwidth:   bandwidthP,
					BillingType: billingType,
					BillingUnit: billingUnit,
					Price:       price,
					Currency:    "USD",
				},
			}
		} else {
			conn = &abi.DoDSettleChangeConnectionParam{
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					ItemId:      fmt.Sprintf("itemid%d", rand.Int31n(100)),
					Bandwidth:   bandwidthP,
					BillingType: billingType,
					StartTime:   startTime,
					EndTime:     endTime,
					Price:       price,
					Currency:    "USD",
				},
			}
		}

		conn.ProductId = productId
		conn.QuoteId = fmt.Sprintf("quote%d", rand.Int())
		conn.QuoteItemId = fmt.Sprintf("quoteItem%d", rand.Int())
		param.Connections = append(param.Connections, conn)
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getChangeOrderBlock", param)
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
