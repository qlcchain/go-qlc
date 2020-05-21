package abi

import (
	"strings"
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/vm/abi"
)

func addDoDSettleTestOrder(t *testing.T, ctx *vmstore.VMContext, order *DoDSettleOrderInfo, id types.Hash) {
	data, err := order.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, DoDSettleDBTableOrder)
	key = append(key, id.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	orderKey := &DoDSettleOrder{
		Seller:  order.Seller.Address,
		OrderId: order.OrderId,
	}

	key = key[0:0]
	key = append(key, DoDSettleDBTableOrderIdMap)
	key = append(key, orderKey.Hash().Bytes()...)

	err = ctx.SetStorage(nil, key, id.Bytes())
	if err != nil {
		t.Fatal(err)
	}
}

func addDoDSettleTestConnection(t *testing.T, ctx *vmstore.VMContext, conn *DoDSettleConnectionInfo, seller types.Address) {
	productKey := &DoDSettleProduct{
		Seller:    seller,
		ProductId: conn.ProductId,
	}
	productHash := productKey.Hash()

	data, err := conn.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, DoDSettleDBTableProduct)
	key = append(key, productHash.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettlementABI(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonDoDSettlement))
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleBillingUnitRound(t *testing.T) {
	start := time.Now().Unix() - 10
	to := start + 1

	end := DoDSettleBillingUnitRound(DoDSettleBillingUnitYear, start, start)
	if end != start {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitNull, start, to)
	if end != to {
		t.Fatal()
	}

	end = DoDSettleBillingUnitRound(DoDSettleBillingUnitYear, start, to)
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

	if DoDSettleCalcBillingUnit(DoDSettleBillingUnitNull, start, end) != 1 {
		t.Fatal()
	}
}

func TestDodSettleCalcAmount(t *testing.T) {
	var bs, be, s, e int64
	dc := new(DoDSettleInvoiceConnDynamic)
	price := 8.0000

	bs = 3600 * 2
	be = 3600 * 10
	s = 0
	e = 3600 * 7
	cp := DoDSettleCalcAmount(bs, be, s, e, price, dc)
	if cp != 5 {
		t.Fatal()
	}

	s = 3600 * 4
	e = 3660 * 15
	cp = DoDSettleCalcAmount(bs, be, s, e, price, dc)
	if cp != 6 {
		t.Fatal()
	}

	s = 3600 * 7
	e = 3600 * 9
	cp = DoDSettleCalcAmount(bs, be, s, e, price, dc)
	if cp != 2 {
		t.Fatal()
	}
}

func TestDoDSettleGetOrderInfoByInternalId(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	order := new(DoDSettleOrderInfo)
	order.OrderId = "order001"
	order.Seller = &DoDSettleUser{Address: mock.Address(), Name: "S1"}
	internalId := mock.Hash()

	_, err := DoDSettleGetOrderInfoByInternalId(ctx, internalId)
	if err == nil {
		t.Fatal()
	}

	_, err = DoDSettleGetInternalIdByOrderId(ctx, order.Seller.Address, order.OrderId)
	if err == nil {
		t.Fatal()
	}

	_, err = DoDSettleGetOrderInfoByOrderId(ctx, order.Seller.Address, order.OrderId)
	if err == nil {
		t.Fatal()
	}

	addDoDSettleTestOrder(t, ctx, order, internalId)

	oi, err := DoDSettleGetOrderInfoByInternalId(ctx, internalId)
	if err != nil || oi.OrderId != order.OrderId {
		t.Fatal()
	}

	id, err := DoDSettleGetInternalIdByOrderId(ctx, order.Seller.Address, order.OrderId)
	if err != nil || id != internalId {
		t.Fatal()
	}

	ori, err := DoDSettleGetOrderInfoByOrderId(ctx, order.Seller.Address, order.OrderId)
	if err != nil || ori.OrderId != order.OrderId {
		t.Fatal()
	}
}

func TestDoDSettleGetConnectionInfoByProductHash(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	seller := mock.Address()
	conn := new(DoDSettleConnectionInfo)
	conn.ProductId = "product001"

	productKey := &DoDSettleProduct{Seller: seller, ProductId: conn.ProductId}
	productHash := productKey.Hash()

	_, err := DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
	if err == nil {
		t.Fatal()
	}

	_, err = DoDSettleGetConnectionInfoByProductId(ctx, seller, conn.ProductId)
	if err == nil {
		t.Fatal()
	}

	addDoDSettleTestConnection(t, ctx, conn, seller)

	ci, err := DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
	if err != nil || ci.ProductId != conn.ProductId {
		t.Fatal()
	}

	cii, err := DoDSettleGetConnectionInfoByProductId(ctx, seller, conn.ProductId)
	if err != nil || cii.ProductId != conn.ProductId {
		t.Fatal()
	}
}

func TestDoDSettleUpdateOrder(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	order := new(DoDSettleOrderInfo)

	err := DoDSettleUpdateOrder(ctx, order, mock.Hash())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleUpdateConnection(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	conn := new(DoDSettleConnectionInfo)

	err := DoDSettleUpdateConnection(ctx, conn, mock.Hash())
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleUpdateConnectionRawParam(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	id := mock.Hash()
	cp1 := new(DoDSettleConnectionParam)
	cp1.ConnectionName = "conn1"
	cp1.PaymentType = DoDSettlePaymentTypeInvoice
	cp1.BillingType = DoDSettleBillingTypePAYG
	cp1.Currency = "USD"
	cp1.ServiceClass = DoDSettleServiceClassGold
	cp1.Bandwidth = "100 Mbps"
	cp1.BillingUnit = DoDSettleBillingUnitHour
	cp1.Price = 10
	cp1.StartTime = 100
	cp1.EndTime = 1000

	err := DoDSettleUpdateConnectionRawParam(ctx, cp1, id)
	if err != nil {
		t.Fatal(err)
	}

	crp1, err := DoDSettleGetConnectionRawParam(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if crp1.ConnectionName != cp1.ConnectionName || crp1.PaymentType != cp1.PaymentType || crp1.BillingType != cp1.BillingType ||
		crp1.Currency != cp1.Currency || crp1.ServiceClass != cp1.ServiceClass || crp1.Bandwidth != cp1.Bandwidth ||
		crp1.Price != cp1.Price || crp1.StartTime != cp1.StartTime || crp1.EndTime != cp1.EndTime {
		t.Fatal()
	}

	cp2 := new(DoDSettleConnectionParam)
	cp2.ConnectionName = "conn2"
	cp2.PaymentType = DoDSettlePaymentTypeStableCoin
	cp2.BillingType = DoDSettleBillingTypeDOD
	cp2.Currency = "CNY"
	cp2.ServiceClass = DoDSettleServiceClassSilver
	cp2.Bandwidth = "200 Mbps"
	cp2.BillingUnit = DoDSettleBillingUnitSecond
	cp2.Price = 100
	cp2.StartTime = 1000
	cp2.EndTime = 10000

	err = DoDSettleUpdateConnectionRawParam(ctx, cp2, id)
	if err != nil {
		t.Fatal(err)
	}

	crp2, err := DoDSettleGetConnectionRawParam(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if crp2.ConnectionName != cp2.ConnectionName || crp2.PaymentType != cp2.PaymentType || crp2.BillingType != cp2.BillingType ||
		crp2.Currency != cp2.Currency || crp2.ServiceClass != cp2.ServiceClass || crp2.Bandwidth != cp2.Bandwidth ||
		crp2.Price != cp2.Price || crp2.StartTime != cp2.StartTime || crp2.EndTime != cp2.EndTime {
		t.Fatal()
	}
}

func TestDoDSettleInheritParam(t *testing.T) {
	dst := new(DoDSettleConnectionDynamicParam)
	src := &DoDSettleConnectionDynamicParam{
		ConnectionName: "conn1",
		PaymentType:    DoDSettlePaymentTypeStableCoin,
		BillingType:    DoDSettleBillingTypeDOD,
		Currency:       "USD",
		ServiceClass:   DoDSettleServiceClassSilver,
		Bandwidth:      "100 Mbps",
		BillingUnit:    DoDSettleBillingUnitMonth,
	}

	DoDSettleInheritParam(src, dst)

	if dst.ConnectionName != src.ConnectionName || dst.PaymentType != src.PaymentType || dst.BillingType != src.BillingType ||
		dst.Currency != src.Currency || dst.ServiceClass != src.ServiceClass || dst.Bandwidth != src.Bandwidth ||
		dst.BillingUnit != src.BillingUnit {
		t.Fatal()
	}
}

func TestDoDSettleGetInternalIdListByAddress(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	buyer := mock.Address()
	seller := mock.Address()

	var key []byte
	key = append(key, DoDSettleDBTableUser)
	key = append(key, buyer.Bytes()...)

	internalId1 := mock.Hash()
	internalId2 := mock.Hash()
	product1 := "product1"
	product2 := "product2"
	order1 := "order1"
	order2 := "order2"

	_, err := DoDSettleGetInternalIdListByAddress(ctx, buyer)
	if err == nil {
		t.Fatal()
	}

	_, err = DoDSettleGetOrderIdListByAddress(ctx, buyer)
	if err == nil {
		t.Fatal()
	}

	_, err = DoDSettleGetProductIdListByAddress(ctx, buyer)
	if err == nil {
		t.Fatal()
	}

	userInfo := new(DoDSettleUserInfos)
	userInfo.InternalIds = []*DoDSettleInternalIdWrap{{InternalId: internalId1}, {InternalId: internalId2}}
	userInfo.ProductIds = []*DoDSettleProduct{{Seller: seller, ProductId: product1}, {Seller: seller, ProductId: product2}}
	userInfo.OrderIds = []*DoDSettleOrder{{Seller: seller, OrderId: order1}, {Seller: seller, OrderId: order2}}

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	il, err := DoDSettleGetInternalIdListByAddress(ctx, buyer)
	if err != nil {
		t.Fatal(err)
	}

	var internalId1Get, internalId2Get bool
	for _, id := range il {
		if id == internalId1 {
			internalId1Get = true
		}

		if id == internalId2 {
			internalId2Get = true
		}
	}
	if !internalId1Get || !internalId2Get {
		t.Fatal()
	}

	ol, err := DoDSettleGetOrderIdListByAddress(ctx, buyer)
	if err != nil {
		t.Fatal(err)
	}

	var orderId1Get, orderId2Get bool
	for _, id := range ol {
		if id.OrderId == order1 {
			orderId1Get = true
		}

		if id.OrderId == order2 {
			orderId2Get = true
		}
	}
	if !orderId1Get || !orderId2Get {
		t.Fatal()
	}

	pl, err := DoDSettleGetProductIdListByAddress(ctx, buyer)
	if err != nil {
		t.Fatal(err)
	}

	var productId1Get, productId2Get bool
	for _, id := range pl {
		if id.ProductId == product1 {
			productId1Get = true
		}

		if id.ProductId == product2 {
			productId2Get = true
		}
	}
	if !productId1Get || !productId2Get {
		t.Fatal()
	}
}

func TestDodSettleGenerateInvoiceByOrder(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	buyer := mock.Address()
	seller := mock.Address()

	order := &DoDSettleOrderInfo{
		Buyer: &DoDSettleUser{
			Address: buyer,
			Name:    "B1",
		},
		Seller: &DoDSettleUser{
			Address: seller,
			Name:    "S1",
		},
		OrderId:       "order001",
		OrderType:     DoDSettleOrderTypeCreate,
		OrderState:    DoDSettleOrderStateSuccess,
		ContractState: DoDSettleContractStateConfirmed,
		Connections:   make([]*DoDSettleConnectionParam, 0),
	}

	conn1 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product001"},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypeDOD,
			Currency:    "USD",
			Price:       8.000,
			StartTime:   1000,
			EndTime:     9000,
		},
	}
	connInfo1 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product001"},
		Active: &DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypeDOD,
			Currency:    "USD",
			Price:       8.000,
			Addition:    8.000,
			StartTime:   1000,
			EndTime:     9000,
		},
	}
	conn2 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product002"},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypeDOD,
			Currency:    "USD",
			Price:       4.000,
			StartTime:   1000,
			EndTime:     5000,
		},
	}
	connInfo2 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product002"},
		Disconnect: &DoDSettleDisconnectInfo{
			OrderId:      "order001",
			Price:        0,
			Currency:     "USD",
			DisconnectAt: 3000,
		},
	}
	conn3 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product003"},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypePAYG,
			BillingUnit: DoDSettleBillingUnitHour,
			Currency:    "USD",
			Price:       4.000,
			StartTime:   1000,
		},
	}
	connInfo3 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product003"},
		Active: &DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypePAYG,
			BillingUnit: DoDSettleBillingUnitHour,
			Currency:    "USD",
			Price:       4.000,
			StartTime:   1000,
		},
	}
	conn4 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product004"},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypePAYG,
			BillingUnit: DoDSettleBillingUnitHour,
			Currency:    "USD",
			Price:       4.000,
			StartTime:   3600,
			EndTime:     7200,
		},
	}
	connInfo4 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product004"},
		Done: []*DoDSettleConnectionDynamicParam{
			{
				OrderId:     "order001",
				BillingType: DoDSettleBillingTypePAYG,
				BillingUnit: DoDSettleBillingUnitHour,
				Currency:    "USD",
				Price:       4.000,
				StartTime:   3600,
				EndTime:     7200,
			},
		},
	}

	order.Connections = append(order.Connections, conn1, conn2, conn3, conn4)
	internalId := mock.Hash()
	addDoDSettleTestOrder(t, ctx, order, internalId)

	addDoDSettleTestConnection(t, ctx, connInfo1, seller)
	addDoDSettleTestConnection(t, ctx, connInfo2, seller)
	addDoDSettleTestConnection(t, ctx, connInfo3, seller)
	addDoDSettleTestConnection(t, ctx, connInfo4, seller)

	start := int64(500)
	end := int64(8000)
	invoice, err := DodSettleGenerateInvoiceByOrder(ctx, seller, order.OrderId, start, end, true, true)
	if err != nil {
		t.Fatal(err)
	}

	if invoice.TotalAmount != float64(19) {
		t.Fatal()
	}

	start = int64(2000)
	end = int64(4000)
	invoice, err = DodSettleGenerateInvoiceByOrder(ctx, seller, order.OrderId, start, end, true, false)
	if err != nil {
		t.Fatal(err)
	}

	if invoice.TotalAmount != float64(4) {
		t.Fatal()
	}

	start = int64(4000)
	end = int64(10000)
	invoice, err = DodSettleGenerateInvoiceByOrder(ctx, seller, order.OrderId, start, end, true, true)
	if err != nil {
		t.Fatal(err)
	}

	if invoice.TotalAmount != 13 {
		t.Fatal()
	}

	start = int64(4000)
	end = int64(0)
	invoice, err = DodSettleGenerateInvoiceByOrder(ctx, seller, order.OrderId, start, end, true, false)
	if err != nil {
		t.Fatal(err)
	}

	start = int64(0)
	end = int64(1000)
	invoice, err = DodSettleGenerateInvoiceByOrder(ctx, seller, order.OrderId, start, end, false, true)
	if err == nil {
		t.Fatal(err)
	}
}

func TestDodSettleGenerateInvoiceByProduct(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	buyer := mock.Address()
	seller := mock.Address()

	order := &DoDSettleOrderInfo{
		Buyer: &DoDSettleUser{
			Address: buyer,
			Name:    "B1",
		},
		Seller: &DoDSettleUser{
			Address: seller,
			Name:    "S1",
		},
		OrderId:       "order001",
		OrderType:     DoDSettleOrderTypeCreate,
		OrderState:    DoDSettleOrderStateSuccess,
		ContractState: DoDSettleContractStateConfirmed,
		Connections: []*DoDSettleConnectionParam{
			{
				DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{Currency: "USD"},
			},
		},
	}

	internalId := mock.Hash()
	addDoDSettleTestOrder(t, ctx, order, internalId)

	conn1 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product001"},
		Active: &DoDSettleConnectionDynamicParam{
			OrderId:     "order003",
			BillingType: DoDSettleBillingTypeDOD,
			Currency:    "USD",
			Price:       5.000,
			Addition:    5.000,
			StartTime:   10000,
			EndTime:     15000,
		},
		Done: []*DoDSettleConnectionDynamicParam{
			{
				OrderId:     "order001",
				BillingType: DoDSettleBillingTypeDOD,
				Currency:    "USD",
				Price:       4.000,
				Addition:    4.000,
				StartTime:   1000,
				EndTime:     5000,
			},
			{
				OrderId:     "order002",
				BillingType: DoDSettleBillingTypeDOD,
				Currency:    "USD",
				Price:       3.000,
				Addition:    3.000,
				StartTime:   6000,
				EndTime:     9000,
			},
		},
		Disconnect: &DoDSettleDisconnectInfo{
			OrderId:      "order004",
			Price:        0,
			Currency:     "USD",
			DisconnectAt: 10000,
		},
		Track: []*DoDSettleConnectionLifeTrack{
			{
				OrderType: order.OrderType,
				OrderId:   order.OrderId,
				Time:      time.Now().Unix(),
			},
		},
	}

	conn2 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product002"},
		Active: &DoDSettleConnectionDynamicParam{
			OrderId:     "order003",
			BillingType: DoDSettleBillingTypePAYG,
			BillingUnit: DoDSettleBillingUnitSecond,
			Currency:    "USD",
			Price:       5.000,
			StartTime:   10000,
		},
		Done: []*DoDSettleConnectionDynamicParam{
			{
				OrderId:     "order001",
				BillingType: DoDSettleBillingTypePAYG,
				BillingUnit: DoDSettleBillingUnitSecond,
				Currency:    "USD",
				Price:       1.000,
				StartTime:   1000,
				EndTime:     5000,
			},
			{
				OrderId:     "order002",
				BillingType: DoDSettleBillingTypePAYG,
				BillingUnit: DoDSettleBillingUnitSecond,
				Currency:    "USD",
				Price:       1.000,
				StartTime:   6000,
				EndTime:     9000,
			},
		},
		Track: []*DoDSettleConnectionLifeTrack{
			{
				OrderType: order.OrderType,
				OrderId:   order.OrderId,
				Time:      time.Now().Unix(),
			},
		},
	}

	addDoDSettleTestConnection(t, ctx, conn1, seller)
	addDoDSettleTestConnection(t, ctx, conn2, seller)

	start := int64(3000)
	end := int64(12000)
	invoice, err := DodSettleGenerateInvoiceByProduct(ctx, seller, conn1.ProductId, start, end, true, true)
	if err != nil {
		t.Fatal(err)
	}

	if invoice.TotalAmount != float64(7) {
		t.Fatal()
	}

	start = int64(0)
	end = int64(12000)
	_, err = DodSettleGenerateInvoiceByProduct(ctx, seller, conn1.ProductId, start, end, true, true)
	if err == nil {
		t.Fatal(err)
	}

	start = int64(3000)
	end = int64(0)
	_, err = DodSettleGenerateInvoiceByProduct(ctx, seller, conn1.ProductId, start, end, true, true)
	if err != nil {
		t.Fatal(err)
	}

	start = int64(12000)
	end = int64(15000)
	invoice, err = DodSettleGenerateInvoiceByProduct(ctx, seller, conn2.ProductId, start, end, true, true)
	if err != nil {
		t.Fatal(err)
	}

	start = int64(3000)
	end = int64(15000)
	invoice, err = DodSettleGenerateInvoiceByProduct(ctx, seller, conn2.ProductId, start, end, true, true)
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println(util.ToIndentString(invoice))
}

func TestDodSettleGenerateInvoiceByBuyer(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	buyer := mock.Address()
	seller := mock.Address()

	order := &DoDSettleOrderInfo{
		Buyer: &DoDSettleUser{
			Address: buyer,
			Name:    "B1",
		},
		Seller: &DoDSettleUser{
			Address: seller,
			Name:    "S1",
		},
		OrderId:       "order001",
		OrderType:     DoDSettleOrderTypeCreate,
		OrderState:    DoDSettleOrderStateSuccess,
		ContractState: DoDSettleContractStateConfirmed,
		Connections:   make([]*DoDSettleConnectionParam, 0),
	}

	conn1 := &DoDSettleConnectionParam{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product001"},
		DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypeDOD,
			Currency:    "USD",
			Price:       8.000,
			StartTime:   1000,
			EndTime:     9000,
		},
	}
	connInfo1 := &DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ProductId: "product001"},
		Active: &DoDSettleConnectionDynamicParam{
			OrderId:     "order001",
			BillingType: DoDSettleBillingTypeDOD,
			Currency:    "USD",
			Price:       8.000,
			StartTime:   1000,
			EndTime:     9000,
		},
	}

	order.Connections = append(order.Connections, conn1)
	internalId := mock.Hash()
	addDoDSettleTestOrder(t, ctx, order, internalId)
	addDoDSettleTestConnection(t, ctx, connInfo1, seller)

	userInfo := new(DoDSettleUserInfos)
	userInfo.OrderIds = []*DoDSettleOrder{{Seller: seller, OrderId: order.OrderId}}

	var key []byte
	key = append(key, DoDSettleDBTableUser)
	key = append(key, buyer.Bytes()...)

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	start := int64(0)
	end := int64(8000)
	_, err = DodSettleGenerateInvoiceByBuyer(ctx, seller, buyer, start, end, true, true)
	if err == nil {
		t.Fatal()
	}

	start = int64(1000)
	end = int64(8000)
	_, err = DodSettleGenerateInvoiceByBuyer(ctx, seller, buyer, start, end, true, true)
	if err != nil {
		t.Fatal()
	}
}

func TestDodSettleGetSellerConnectionActive(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	active := new(DoDSettleConnectionActive)
	active.ActiveAt = time.Now().Unix()
	id := mock.Hash()

	err := DodSettleSetSellerConnectionActive(ctx, active, id)
	if err != nil {
		t.Fatal()
	}

	ac, err := DodSettleGetSellerConnectionActive(ctx, id)
	if err != nil || ac.ActiveAt != active.ActiveAt {
		t.Fatal(err)
	}
}
