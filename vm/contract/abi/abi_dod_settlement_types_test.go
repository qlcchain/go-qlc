package abi

import (
	"bytes"
	"testing"

	"github.com/tinylib/msgp/msgp"

	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/vmstore"
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

	cop.Connections = []*DoDSettleConnectionParam{
		{
			DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
				SrcCompanyName: "scn",
				SrcRegion:      "sr",
				SrcCity:        "sc",
				SrcDataCenter:  "sdc",
				SrcPort:        "sp",
				DstCompanyName: "dcn",
				DstRegion:      "dr",
				DstCity:        "dc",
				DstDataCenter:  "ddc",
			},
			DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{},
		},
		{
			DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
				SrcCompanyName: "scn",
				SrcRegion:      "sr",
				SrcCity:        "sc",
				SrcDataCenter:  "sdc",
				SrcPort:        "sp",
				DstCompanyName: "dcn",
				DstRegion:      "dr",
				DstCity:        "dc",
				DstDataCenter:  "ddc",
				DstPort:        "dp",
			},
			DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{},
		},
	}
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].DstPort = "dp"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].ProductOfferingId = "po0"
	cop.Connections[1].ProductOfferingId = "po1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].ItemId = "item0"
	cop.Connections[1].ItemId = "item0"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[1].ItemId = "item1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].BuyerProductId = "bp0"
	cop.Connections[1].BuyerProductId = "bp0"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[1].BuyerProductId = "bp1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].QuoteId = "quote0"
	cop.Connections[1].QuoteId = "quote1"
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

	cop.Connections[0].BillingType = DoDSettleBillingTypeDOD
	cop.Connections[1].BillingType = DoDSettleBillingTypePAYG
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].PaymentType = DoDSettlePaymentTypeInvoice
	cop.Connections[1].PaymentType = DoDSettlePaymentTypeStableCoin
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].ServiceClass = DoDSettleServiceClassSilver
	cop.Connections[1].ServiceClass = DoDSettleServiceClassGold
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

	cop.Connections[1].BillingUnit = DoDSettleBillingUnitSecond
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].StartTime = 100
	cop.Connections[0].EndTime = 1000
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].Currency = "CNY"
	cop.Connections[0].Currency = "USD"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].Bandwidth = "10 Mbps"
	cop.Connections[0].Bandwidth = "10 Kbps"
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
				DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{BuyerProductId: "b1"},
			},
			{
				DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{BuyerProductId: "b2"},
			},
		},
	}
	addDoDSettleTestOrder(t, ctx, order, uop.InternalId)

	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.OrderId = "order001"
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.InternalId = mock.Hash()
	order2 := &DoDSettleOrderInfo{
		Buyer: &DoDSettleUser{
			Address: mock.Address(),
			Name:    "B1",
		},
		Seller: &DoDSettleUser{
			Address: mock.Address(),
			Name:    "S1",
		},
		OrderId:   "order002",
		OrderType: DoDSettleOrderTypeCreate,
		Connections: []*DoDSettleConnectionParam{
			{
				DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{BuyerProductId: "b3"},
			},
			{
				DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{BuyerProductId: "b4"},
			},
		},
	}
	err = DoDSettleUpdateOrder(ctx, order2, uop.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	conn := new(DoDSettleConnectionInfo)
	conn.ProductId = "p1"
	addDoDSettleTestConnection(t, ctx, conn, order2.Seller.Address)

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "", BuyerProductId: "b3"}}
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p2", BuyerProductId: ""}}
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p2", BuyerProductId: "b3"}, {ProductId: "p2", BuyerProductId: "b4"}}
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p1", BuyerProductId: "b3"}, {ProductId: "p1", BuyerProductId: "b3"}}
	err = uop.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	uop.ProductIds = []*DoDSettleProductItem{{ProductId: "p2", BuyerProductId: "b3"}, {ProductId: "p3", BuyerProductId: "b4"}}
	err = uop.Verify(ctx)
	if err != nil {
		t.Fatal(err)
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

	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections = []*DoDSettleChangeConnectionParam{{}, {}}
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].ProductId = "product0"
	cop.Connections[1].ProductId = "product1"
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].QuoteId = "quote0"
	cop.Connections[1].QuoteId = "quote1"
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
	cop.Connections[1].BillingType = DoDSettleBillingTypeDOD
	err = cop.Verify()
	if err == nil {
		t.Fatal()
	}

	cop.Connections[0].StartTime = 10
	cop.Connections[0].EndTime = 20
	cop.Connections[1].StartTime = 10
	cop.Connections[1].EndTime = 20
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
	err = top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	top.Connections = []*DoDSettleChangeConnectionParam{{}}
	err = top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	top.Connections[0].ProductId = "p1"
	err = top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	conn := new(DoDSettleConnectionInfo)
	conn.ProductId = "p1"
	conn.Active = &DoDSettleConnectionDynamicParam{BillingType: DoDSettleBillingTypeDOD}
	addDoDSettleTestConnection(t, ctx, conn, top.Seller.Address)

	err = top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	top.Connections[0].QuoteId = "quote0"
	err = top.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	top.Connections[0].QuoteItemId = "quoteItem0"
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
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	rrp := new(DoDSettleUpdateProductInfoParam)

	err := rrp.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	rrp.Address = mock.Address()
	rrp.InternalId = mock.Hash()
	err = rrp.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	rrp.ProductId = []string{"p1"}
	err = rrp.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	order := NewOrderInfo()
	order.Connections = []*DoDSettleConnectionParam{{}}
	order.Connections[0].ProductId = "p2"
	err = DoDSettleUpdateOrder(ctx, order, rrp.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	err = rrp.Verify(ctx)
	if err == nil {
		t.Fatal()
	}

	rrp.ProductId = []string{"p2"}
	err = rrp.Verify(ctx)
	if err != nil {
		t.Fatal()
	}

	ak := &DoDSettleConnectionActiveKey{InternalId: rrp.InternalId, ProductId: "p2"}
	err = DoDSettleSetSellerConnectionActive(ctx, &DoDSettleConnectionActive{ActiveAt: 1111}, ak.Hash())
	if err != nil {
		t.Fatal()
	}

	err = rrp.Verify(ctx)
	if err == nil {
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

func TestDoDSettleConnectionActiveKeyHash(t *testing.T) {
	ak := &DoDSettleConnectionActiveKey{InternalId: mock.Hash(), ProductId: "p1"}
	_ = ak.Hash()
}

func TestMarshalUnmarshalDoDSettleContractState(t *testing.T) {
	v := DoDSettleContractState(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleContractState(t *testing.T) {
	v := DoDSettleContractState(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleContractState(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettleOrderState(t *testing.T) {
	v := DoDSettleOrderState(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleOrderState(t *testing.T) {
	v := DoDSettleOrderState(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleOrderState(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettlePaymentType(t *testing.T) {
	v := DoDSettlePaymentType(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettlePaymentType(t *testing.T) {
	v := DoDSettlePaymentType(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettlePaymentType(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettleBillingUnit(t *testing.T) {
	v := DoDSettleBillingUnit(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleBillingUnit(t *testing.T) {
	v := DoDSettleBillingUnit(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleBillingUnit(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettleServiceClass(t *testing.T) {
	v := DoDSettleServiceClass(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleServiceClass(t *testing.T) {
	v := DoDSettleServiceClass(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleServiceClass(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettleResponseAction(t *testing.T) {
	v := DoDSettleResponseAction(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleResponseAction(t *testing.T) {
	v := DoDSettleResponseAction(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleResponseAction(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettleBillingType(t *testing.T) {
	v := DoDSettleBillingType(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleBillingType(t *testing.T) {
	v := DoDSettleBillingType(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleBillingType(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}

func TestMarshalUnmarshalDoDSettleOrderType(t *testing.T) {
	v := DoDSettleOrderType(1)
	bts, err := v.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}
	left, err := v.UnmarshalMsg(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after UnmarshalMsg(): %q", len(left), left)
	}

	left, err = msgp.Skip(bts)
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("%d bytes left over after Skip(): %q", len(left), left)
	}
}

func TestEncodeDecodeDoDSettleOrderType(t *testing.T) {
	v := DoDSettleOrderType(1)
	var buf bytes.Buffer
	msgp.Encode(&buf, &v)

	m := v.Msgsize()
	if buf.Len() > m {
		t.Log("WARNING: Msgsize() is inaccurate")
	}

	vn := DoDSettleOrderType(1)
	err := msgp.Decode(&buf, &vn)
	if err != nil {
		t.Error(err)
	}

	buf.Reset()
	msgp.Encode(&buf, &v)
	err = msgp.NewReader(&buf).Skip()
	if err != nil {
		t.Error(err)
	}
}
