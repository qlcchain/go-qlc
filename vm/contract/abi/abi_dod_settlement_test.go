package abi

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/vm/abi"
)

func TestDoDSettlementABI(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonDoDSettlement))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleBillingUnitRound(t *testing.T) {
	start := time.Now().Unix() - 10
	to := start + 1

	end := DoDSettleBillingUnitRound(DoDSettleBillingUnitYear, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitYear, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitMonth, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitMonth, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitWeek, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitWeek, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitDay, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitDay, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitHour, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitHour, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitMinute, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitMinute, start, end) != 1 {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitSecond, start, to)
	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitSecond, start, end) != 1 {
		t.Fatal()
	}
}

func TestDodSettleGenerateInvoiceByOrder(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)

	order := DoDSettleOrderInfo{
		Buyer: &DoDSettleUser{
			Address: mock.Address(),
			Name:    "B1",
		},
		Seller: &DoDSettleUser{
			Address: mock.Address(),
			Name:    "S1",
		},
		OrderId:       "order001",
		OrderType:     DoDSettleOrderTypeCreate,
		OrderState:    0,
		ContractState: DoDSettleContractStateConfirmed,
		Connections:   make([]*DoDSettleConnectionParam, 0),
		Track:         nil,
	}

	conn1 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
			ProductId: "product001",
		},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			Currency: "USD",
		},
	}
	conn2 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
			ProductId: "product002",
		},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			Currency: "USD",
		},
	}
	order.Connections = append(order.Connections, conn1, conn2)

	data, err := order.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	internalId := mock.Hash()
	var key []byte
	key = append(key, DoDSettleDBTableOrder)
	key = append(key, internalId.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	orderKey := &DoDSettleOrder{Seller: order.Seller.Address, OrderId: order.OrderId}
	key = key[0:0]
	key = append(key, DoDSettleDBTableOrderIdMap)
	key = append(key, orderKey.Hash().Bytes()...)

	err = ctx.SetStorage(nil, key, internalId.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now().Unix() - 5000

	connInfo1 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
			ProductId: conn1.ProductId,
		},
		Active: &DoDSettleConnectionDynamicParam{
			PaymentType:  DoDSettlePaymentTypeInvoice,
			BillingType:  DoDSettleBillingTypeDOD,
			Currency:     "USD",
			ServiceClass: DoDSettleServiceClassGold,
			Bandwidth:    "1000 Mbps",
			BillingUnit:  DoDSettleBillingUnitSecond,
			Price:        100,
			StartTime:    start + 300,
			EndTime:      start + 400,
		},
		Done: make([]*DoDSettleConnectionDynamicParam, 0),
	}
	ciDone1 := &DoDSettleConnectionDynamicParam{
		PaymentType:  DoDSettlePaymentTypeInvoice,
		BillingType:  DoDSettleBillingTypeDOD,
		Currency:     "USD",
		ServiceClass: DoDSettleServiceClassGold,
		Bandwidth:    "1000 Mbps",
		BillingUnit:  DoDSettleBillingUnitSecond,
		Price:        300,
		StartTime:    start,
		EndTime:      start + 300,
	}
	connInfo1.Done = append(connInfo1.Done, ciDone1)

	data, err = connInfo1.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	productKey := &DoDSettleProduct{Seller: order.Seller.Address, ProductId: connInfo1.ProductId}
	key = key[0:0]
	key = append(key, DoDSettleDBTableProduct)
	key = append(key, productKey.Hash().Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	userInfo := new(DoDSettleUserInfos)
	orderId := &DoDSettleOrder{Seller: order.Seller.Address, OrderId: order.OrderId}
	userInfo.OrderIds = []*DoDSettleOrder{orderId}

	data, err = userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	key = key[0:0]
	key = append(key, DoDSettleDBTableUser)
	key = append(key, order.Buyer.Address.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	connInfo2 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
			ProductId: conn2.ProductId,
		},
		Active: &DoDSettleConnectionDynamicParam{
			PaymentType:  DoDSettlePaymentTypeInvoice,
			BillingType:  DoDSettleBillingTypePAYG,
			Currency:     "USD",
			ServiceClass: DoDSettleServiceClassGold,
			Bandwidth:    "1000 Mbps",
			BillingUnit:  DoDSettleBillingUnitHour,
			Price:        500,
			StartTime:    start + 320,
		},
		Done: make([]*DoDSettleConnectionDynamicParam, 0),
	}

	ci2Done1 := &DoDSettleConnectionDynamicParam{
		PaymentType:  DoDSettlePaymentTypeInvoice,
		BillingType:  DoDSettleBillingTypePAYG,
		Currency:     "USD",
		ServiceClass: DoDSettleServiceClassGold,
		Bandwidth:    "1000 Mbps",
		BillingUnit:  DoDSettleBillingUnitMinute,
		Price:        10,
		StartTime:    start - 100,
		EndTime:      start + 200,
	}
	connInfo2.Done = append(connInfo2.Done, ci2Done1)

	ci2Done2 := &DoDSettleConnectionDynamicParam{
		PaymentType:  DoDSettlePaymentTypeInvoice,
		BillingType:  DoDSettleBillingTypePAYG,
		Currency:     "USD",
		ServiceClass: DoDSettleServiceClassGold,
		Bandwidth:    "1000 Mbps",
		BillingUnit:  DoDSettleBillingUnitMinute,
		Price:        10,
		StartTime:    start + 200,
		EndTime:      start + 320,
	}
	connInfo2.Done = append(connInfo2.Done, ci2Done2)

	data, err = connInfo2.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	productKey = &DoDSettleProduct{Seller: order.Seller.Address, ProductId: connInfo2.ProductId}
	key = key[0:0]
	key = append(key, DoDSettleDBTableProduct)
	key = append(key, productKey.Hash().Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	invoice1, err := DodSettleGenerateInvoiceByOrder(ctx, order.Seller.Address, order.OrderId, start, start+1000)
	if err != nil {
		t.Fatal(err)
	}

	if invoice1.TotalAmount != 950 {
		t.Fatal()
	}

	invoice2, err := DodSettleGenerateInvoiceByBuyer(ctx, order.Seller.Address, order.Buyer.Address, start, start+1000)
	if err != nil {
		t.Fatal()
	}

	fmt.Println(util.ToIndentString(invoice2))
}
