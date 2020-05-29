package commands

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/abiosoft/ishell"
	rpc "github.com/qlcchain/jsonrpc2"

	"github.com/qlcchain/go-qlc/cmd/util"
	"github.com/qlcchain/go-qlc/common/types"
	cutil "github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addDSCreateOrderCmdByShell(parentCmd *ishell.Cmd) {
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
	srcPort := util.Flag{
		Name:  "srcPort",
		Must:  true,
		Usage: "source port",
		Value: "",
	}
	dstPort := util.Flag{
		Name:  "dstPort",
		Must:  true,
		Usage: "destination port",
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
	num := util.Flag{
		Name:  "num",
		Must:  true,
		Usage: "num",
		Value: "",
	}

	args := []util.Flag{buyerAddress, buyerName, sellerAddress, sellerName, srcPort, dstPort, billingType, bandwidth,
		billingUnit, price, startTime, endTime, num}
	cmd := &ishell.Cmd{
		Name:                "createOrder",
		Help:                "create a order request",
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
			srcPortP := util.StringVar(c.Args, srcPort)
			dstPortP := util.StringVar(c.Args, dstPort)
			billingTypeP := util.StringVar(c.Args, billingType)
			bandwidthP := util.StringVar(c.Args, bandwidth)
			billingUnitP := util.StringVar(c.Args, billingUnit)
			priceP := util.StringVar(c.Args, price)
			startTimeP := util.StringVar(c.Args, startTime)
			endTimeP := util.StringVar(c.Args, endTime)
			numP := util.StringVar(c.Args, num)

			if err := DSCreateOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, srcPortP, dstPortP,
				billingTypeP, bandwidthP, billingUnitP, priceP, startTimeP, endTimeP, numP); err != nil {
				util.Warn(err)
				return
			}
		},
	}
	parentCmd.AddCmd(cmd)
}

func DSCreateOrder(buyerAddressP, buyerNameP, sellerAddressP, sellerNameP, srcPortP, dstPortP, billingTypeP,
	bandwidthP, billingUnitP, priceP, startTimeP, endTimeP, numP string) error {
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

	paymentType, err := abi.ParseDoDSettlePaymentType("invoice")
	if err != nil {
		return err
	}

	billingType, err := abi.ParseDoDSettleBillingType(billingTypeP)
	if err != nil {
		return err
	}

	var billingUnit abi.DoDSettleBillingUnit
	if len(billingUnitP) > 0 {
		billingUnit, err = abi.ParseDoDSettleBillingUnit(billingUnitP)
		if err != nil {
			return err
		}
	}

	price, err := strconv.ParseFloat(priceP, 64)
	if err != nil {
		return err
	}

	serviceClass, err := abi.ParseDoDSettleServiceClass("gold")
	if err != nil {
		return err
	}

	num, err := strconv.Atoi(numP)
	if err != nil {
		return err
	}

	param := &abi.DoDSettleCreateOrderParam{
		Buyer: &abi.DoDSettleUser{
			Address: acc.Address(),
			Name:    buyerNameP,
		},
		Seller: &abi.DoDSettleUser{
			Address: sellerAddress,
			Name:    sellerNameP,
		},
		Connections: make([]*abi.DoDSettleConnectionParam, 0),
	}

	var conn *abi.DoDSettleConnectionParam
	for i := 0; i < num; i++ {
		if billingType == abi.DoDSettleBillingTypePAYG {
			conn = &abi.DoDSettleConnectionParam{
				DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
					ItemId:         fmt.Sprintf("item%d", rand.Int()),
					BuyerProductId: fmt.Sprintf("buyerProduct%d", rand.Int()),
					SrcCompanyName: "CBC",
					SrcRegion:      "CHN",
					SrcCity:        "HK",
					SrcDataCenter:  "DCX",
					SrcPort:        srcPortP,
					DstCompanyName: "CBC",
					DstRegion:      "USA",
					DstCity:        "NYC",
					DstDataCenter:  "DCY",
					DstPort:        dstPortP,
				},
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					ConnectionName: fmt.Sprintf("connection%d", rand.Int()),
					QuoteId:        fmt.Sprintf("quote%d", rand.Int()),
					QuoteItemId:    fmt.Sprintf("quoteItem%d", rand.Int()),
					Bandwidth:      bandwidthP,
					BillingUnit:    billingUnit,
					Price:          price,
					ServiceClass:   serviceClass,
					PaymentType:    paymentType,
					BillingType:    billingType,
					Currency:       "USD",
				},
			}
		} else {
			startTime, err := strconv.ParseInt(startTimeP, 10, 64)
			if err != nil {
				return err
			}

			endTime, err := strconv.ParseInt(endTimeP, 10, 64)
			if err != nil {
				return err
			}

			conn = &abi.DoDSettleConnectionParam{
				DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
					ItemId:         fmt.Sprintf("item%d", rand.Int()),
					BuyerProductId: fmt.Sprintf("buyerProduct%d", rand.Int()),
					SrcCompanyName: "CBC",
					SrcRegion:      "CHN",
					SrcCity:        "HK",
					SrcDataCenter:  "DCX",
					SrcPort:        srcPortP,
					DstCompanyName: "CBC",
					DstRegion:      "USA",
					DstCity:        "NYC",
					DstDataCenter:  "DCY",
					DstPort:        dstPortP,
				},
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					ConnectionName: fmt.Sprintf("connection%d", rand.Int()),
					QuoteId:        fmt.Sprintf("quote%d", rand.Int()),
					QuoteItemId:    fmt.Sprintf("quoteItem%d", rand.Int()),
					Bandwidth:      bandwidthP,
					Price:          price,
					ServiceClass:   serviceClass,
					PaymentType:    paymentType,
					BillingType:    billingType,
					Currency:       "USD",
					StartTime:      startTime,
					EndTime:        endTime,
				},
			}
		}

		param.Connections = append(param.Connections, conn)
	}

	block := new(types.StateBlock)
	err = client.Call(&block, "DoDSettlement_getCreateOrderBlock", param)
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
