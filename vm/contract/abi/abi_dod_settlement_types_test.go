package abi

import (
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
	"testing"
)

func TestDoDSettleCreateOrderParam(t *testing.T) {
	cop := new(DoDSettleCreateOrderParam)

	err := cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Buyer = &DoDSettleUser{Address: mock.Address(), Name: "B1"}
	cop.Seller = &DoDSettleUser{Address: mock.Address(), Name: "S1"}
	cop.Connections = make([]*DoDSettleConnectionParam, 0)
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.QuoteId = "quote001"
	err = cop.Verify()
	if err != nil {
		t.Fatal()
	}

	cop.Connections = []*DoDSettleConnectionParam{
		{
			DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
				ItemId: "item1",
			},
			DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{},
		},
		{
			DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
				ItemId: "item2",
			},
			DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{},
		},
	}
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].QuoteItemId = "quoteItem1"
	cop.Connections[1].QuoteItemId = "quoteItem1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[1].QuoteItemId = "quoteItem2"
	cop.Connections[0].BillingType = DoDSettleBillingTypeDOD
	cop.Connections[0].StartTime = 100
	cop.Connections[0].EndTime = 100
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].StartTime = 100
	cop.Connections[0].EndTime = 1000
	err = cop.Verify()
	if err != nil {
		t.Fatal()
	}

	cop.Connections[1].ItemId = "item1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	data, err := cop.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	err = cop.FromABI(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleUpdateOrderInfoParam(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	uop := new(DoDSettleUpdateOrderInfoParam)

	err := uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p1", ItemId: "i1"}, {ProductId: "p1", ItemId: "i1"}}
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.InternalId = mock.Hash()
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	order := &DoDSettleOrderInfo{
		Buyer: &DoDSettleUser{
			Address: mock.Address(),
			Name:    "B1",
		},
		Seller: &DoDSettleUser{
			Address: mock.Address(),
			Name:    "S1",
		},
		OrderId:   "order001",
		OrderType: DoDSettleOrderTypeCreate,
		Connections: []*DoDSettleConnectionParam{
			{
				DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ItemId: "i1"},
			},
			{
				DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{ItemId: "i2"},
			},
		},
	}
	addDoDSettleTestOrder(t, ctx, order, uop.InternalId)

	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p1", ItemId: "i1"}, {ProductId: "p3", ItemId: "i3"}}
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p1", ItemId: "i1"}, {ProductId: "p2", ItemId: "i2"}}
	err = uop.Verify(ctx)
	if err != nil {
		t.Fatal()
	}

	data, err := uop.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	err = uop.FromABI(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleChangeOrderParam(t *testing.T) {
	cop := new(DoDSettleChangeOrderParam)

	err := cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Buyer = &DoDSettleUser{Address: mock.Address(), Name: "B1"}
	cop.Seller = &DoDSettleUser{Address: mock.Address(), Name: "S1"}
	cop.Connections = []*DoDSettleChangeConnectionParam{
		{},
		{},
	}

	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.QuoteId = "quote1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].ProductId = "product1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].QuoteItemId = "quoteItem1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[1].ProductId = "product2"
	cop.Connections[1].QuoteItemId = "quoteItem1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[1].QuoteItemId = "quoteItem2"
	cop.Connections[0].BillingType = DoDSettleBillingTypeDOD
	cop.Connections[1].BillingType = DoDSettleBillingTypeDOD
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].StartTime = 10
	cop.Connections[1].StartTime = 10
	err = cop.Verify()
	if err != nil {
		t.Fatal()
	}

	data, err := cop.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	err = cop.FromABI(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleTerminateOrderParam(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	top := new(DoDSettleTerminateOrderParam)

	err := top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	top.Buyer = &DoDSettleUser{Address: mock.Address(), Name: "B1"}
	top.Seller = &DoDSettleUser{Address: mock.Address(), Name: "S1"}
	top.ProductId = []string{"p1"}

	err = top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	conn := new(DoDSettleConnectionInfo)
	conn.ProductId = "p1"
	addDoDSettleTestConnection(t, ctx, conn, top.Seller.Address)

	err = top.Verify(ctx)
	if err != nil {
		t.Fatal()
	}

	data, err := top.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	err = top.FromABI(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleResourceReadyParam(t *testing.T) {
	rrp := new(DoDSettleResourceReadyParam)

	err := rrp.Verify()
	if err == nil {
		t.Fatal()
	}

	rrp.Address = mock.Address()
	rrp.InternalId = mock.Hash()
	err = rrp.Verify()
	if err == nil {
		t.Fatal()
	}

	rrp.ProductId = []string{"p1"}
	err = rrp.Verify()
	if err != nil {
		t.Fatal()
	}

	data, err := rrp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	err = rrp.FromABI(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewOrderInfo(t *testing.T) {
	o := NewOrderInfo()
	if o == nil {
		t.Fatal()
	}
}
