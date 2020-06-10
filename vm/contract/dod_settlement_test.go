package contract

import (
	"testing"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func TestDoDSettleCreateOrder_ProcessSend(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	co := new(DoDSettleCreateOrder)

	_, _, err := co.ProcessSend(ctx, block)
	if err != ErrToken {
		t.Fatal(err)
	}

	block.Token = cfg.GasToken()
	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(block.Address)
	am.Tokens[0] = mock.TokenMeta2(block.Address, cfg.GasToken())
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	cp := &abi.DoDSettleCreateOrderParam{
		Seller: &abi.DoDSettleUser{Address: mock.Address(), Name: "S1"},
		Connections: []*abi.DoDSettleConnectionParam{
			{
				DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
					ProductOfferingId: "po1",
					SrcCompanyName:    "CBC",
					SrcRegion:         "CHN",
					SrcCity:           "HK",
					SrcDataCenter:     "DCX",
					SrcPort:           "sp001",
					DstCompanyName:    "CBC",
					DstRegion:         "USA",
					DstCity:           "NYC",
					DstDataCenter:     "DCY",
					DstPort:           "dp001",
				},
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					ItemId:         "item1",
					ConnectionName: "conn1",
					QuoteId:        "quote1",
					QuoteItemId:    "quoteItem1",
					Bandwidth:      "200 Mbps",
					BillingUnit:    abi.DoDSettleBillingUnitSecond,
					Price:          1,
					ServiceClass:   abi.DoDSettleServiceClassSilver,
					PaymentType:    abi.DoDSettlePaymentTypeStableCoin,
					BillingType:    abi.DoDSettleBillingTypePAYG,
					Currency:       "USD",
				},
			},
			{
				DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
					ProductOfferingId: "po2",
					SrcCompanyName:    "CBC",
					SrcRegion:         "CHN",
					SrcCity:           "HK",
					SrcDataCenter:     "DCX",
					SrcPort:           "sp001",
					DstCompanyName:    "CBC",
					DstRegion:         "USA",
					DstCity:           "NYC",
					DstDataCenter:     "DCY",
					DstPort:           "dp001",
				},
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					ItemId:         "item2",
					ConnectionName: "conn2",
					QuoteId:        "quote1",
					QuoteItemId:    "quoteItem2",
					Bandwidth:      "200 Mbps",
					Price:          1,
					ServiceClass:   abi.DoDSettleServiceClassSilver,
					PaymentType:    abi.DoDSettlePaymentTypeStableCoin,
					BillingType:    abi.DoDSettleBillingTypeDOD,
					Currency:       "USD",
					StartTime:      1000,
					EndTime:        10000,
				},
			},
		},
	}

	block.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	cp.Buyer = &abi.DoDSettleUser{Address: mock.Address(), Name: "B1"}
	block.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	cp.Buyer = &abi.DoDSettleUser{Address: block.Address, Name: "B1"}
	block.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, cp.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
	userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

	internalId := &abi.DoDSettleInternalIdWrap{InternalId: mock.Hash()}
	userInfo.InternalIds = append(userInfo.InternalIds, internalId)

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleCreateOrder_DoReceive(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	recv := mock.StateBlockWithoutWork()
	co := new(DoDSettleCreateOrder)

	_, err := co.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	param := new(abi.DoDSettleResponseParam)
	param.RequestHash = send.GetHash()
	param.Action = abi.DoDSettleResponseActionConfirm

	recv.Data, err = param.MarshalMsg(nil)
	if err != nil {
		t.Fatal()
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	order := new(abi.DoDSettleOrderInfo)
	order.Seller = &abi.DoDSettleUser{Address: recv.Address, Name: "s1"}
	order.Track = make([]*abi.DoDSettleOrderLifeTrack, 0)
	err = abi.DoDSettleUpdateOrder(ctx, order, send.Previous)
	if err != nil {
		t.Fatal(err)
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	param.Action = abi.DoDSettleResponseActionReject

	recv.Data, err = param.MarshalMsg(nil)
	if err != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(recv.Address)
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	am.Tokens[0] = mock.TokenMeta2(recv.Address, cfg.GasToken())
	err = l.UpdateAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleCreateOrder_GetTargetReceiver(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	co := new(DoDSettleCreateOrder)

	_, err := co.GetTargetReceiver(ctx, send)
	if err == nil {
		t.Fatal()
	}

	cp := &abi.DoDSettleCreateOrderParam{
		Seller: &abi.DoDSettleUser{Address: mock.Address(), Name: "s1"},
	}

	send.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal()
	}

	tr, err := co.GetTargetReceiver(ctx, send)
	if err != nil || tr != cp.Seller.Address {
		t.Fatal()
	}
}

func TestDoDSettleUpdateOrderInfo_ProcessSend(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	uo := new(DoDSettleUpdateOrderInfo)

	_, _, err := uo.ProcessSend(ctx, block)
	if err != ErrToken {
		t.Fatal(err)
	}

	block.Token = cfg.GasToken()
	_, _, err = uo.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(block.Address)
	am.Tokens[0] = mock.TokenMeta2(block.Address, cfg.GasToken())
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	up := &abi.DoDSettleUpdateOrderInfoParam{
		Buyer:   mock.Address(),
		OrderId: "order1",
		Status:  abi.DoDSettleOrderStateSuccess,
	}

	block.Data, err = up.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	up.InternalId = mock.Hash()
	up.OrderItemId = []*abi.DoDSettleOrderItem{{ItemId: "i1", OrderItemId: "oi1"}}
	block.Data, err = up.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	order := abi.NewOrderInfo()
	order.Seller = &abi.DoDSettleUser{Address: mock.Address()}
	order.OrderId = "order1"
	order.OrderType = abi.DoDSettleOrderTypeCreate
	conn := new(abi.DoDSettleConnectionParam)
	conn.ItemId = "i1"
	conn.BillingType = abi.DoDSettleBillingTypeDOD
	order.Connections = append(order.Connections, conn)
	err = abi.DoDSettleUpdateOrder(ctx, order, up.InternalId)
	if err != nil {
		t.Fatal()
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order.Buyer = &abi.DoDSettleUser{Address: block.Address}
	err = abi.DoDSettleUpdateOrder(ctx, order, up.InternalId)
	if err != nil {
		t.Fatal()
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, order.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
	userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	order = abi.NewOrderInfo()
	order.Buyer = &abi.DoDSettleUser{Address: block.Address}
	order.Seller = &abi.DoDSettleUser{Address: mock.Address()}
	order.OrderId = "order2"
	order.OrderType = abi.DoDSettleOrderTypeCreate
	conn = new(abi.DoDSettleConnectionParam)
	conn.ItemId = "i1"
	conn.ProductId = "p2"
	conn.BillingType = abi.DoDSettleBillingTypeDOD
	order.Connections = append(order.Connections, conn)
	err = abi.DoDSettleUpdateOrder(ctx, order, up.InternalId)
	if err != nil {
		t.Fatal()
	}

	up.OrderId = "order002"
	block.Data, err = up.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	pid := &abi.DoDSettleProduct{Seller: order.Seller.Address, ProductId: "p2"}
	otp := &abi.DoDSettleOrderToProduct{Seller: order.Seller.Address, OrderId: order.OrderId, OrderItemId: "oi1"}
	err = abi.DoDSettleSetProductStorageKeyByProductId(ctx, otp.Hash(), pid.Hash())
	if err != nil {
		t.Fatal(err)
	}

	err = abi.DoDSettleSetProductIdByStorageKey(ctx, otp.Hash(), "p2", order.Seller.Address)
	if err != nil {
		t.Fatal(err)
	}

	ci := &abi.DoDSettleConnectionInfo{
		Active: &abi.DoDSettleConnectionDynamicParam{
			BillingType: abi.DoDSettleBillingTypePAYG,
			BillingUnit: abi.DoDSettleBillingUnitSecond,
			StartTime:   1000,
		},
		Done:       make([]*abi.DoDSettleConnectionDynamicParam, 0),
		Disconnect: nil,
		Track:      make([]*abi.DoDSettleConnectionLifeTrack, 0),
	}
	err = abi.DoDSettleUpdateConnection(ctx, ci, otp.Hash())
	if err != nil {
		t.Fatal(err)
	}

	order.OrderType = abi.DoDSettleOrderTypeChange
	err = abi.DoDSettleUpdateOrder(ctx, order, up.InternalId)
	if err != nil {
		t.Fatal()
	}

	up.OrderId = "order003"
	block.Data, err = up.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	order.OrderType = abi.DoDSettleOrderTypeTerminate
	err = abi.DoDSettleUpdateOrder(ctx, order, up.InternalId)
	if err != nil {
		t.Fatal()
	}

	up.OrderId = "order004"
	block.Data, err = up.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleUpdateOrderInfo_DoReceive(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	recv := mock.StateBlockWithoutWork()
	uo := new(DoDSettleUpdateOrderInfo)

	_, err := uo.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	param := new(abi.DoDSettleUpdateOrderInfoParam)
	param.InternalId = mock.Hash()
	param.OrderId = "order1"

	send.Data, err = param.ToABI()
	if err != nil {
		t.Fatal()
	}

	_, err = uo.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	order := new(abi.DoDSettleOrderInfo)
	order.OrderState = abi.DoDSettleOrderStateFail
	order.Seller = &abi.DoDSettleUser{Address: mock.Address()}
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	_, err = uo.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	_, err = uo.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal(err)
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	am := mock.AccountMeta(recv.Address)
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = uo.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal(err)
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	am.Tokens[0] = mock.TokenMeta2(recv.Address, cfg.GasToken())
	err = l.UpdateAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = uo.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleUpdateOrderInfo_GetTargetReceiver(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	uo := new(DoDSettleUpdateOrderInfo)

	_, err := uo.GetTargetReceiver(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param := &abi.DoDSettleUpdateOrderInfoParam{
		Buyer: block.Address,
	}

	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, err = uo.GetTargetReceiver(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param.InternalId = mock.Hash()
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, err = uo.GetTargetReceiver(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order := abi.NewOrderInfo()
	order.Seller = &abi.DoDSettleUser{Address: mock.Address()}
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	tr, err := uo.GetTargetReceiver(ctx, block)
	if err != nil || tr != order.Seller.Address {
		t.Fatal()
	}
}

func TestDoDSettleUpdateOrderInfo_DoGap(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	uo := new(DoDSettleUpdateOrderInfo)

	_, _, err := uo.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param := &abi.DoDSettleUpdateOrderInfoParam{
		Buyer: block.Address,
	}

	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param.InternalId = mock.Hash()
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = uo.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order := abi.NewOrderInfo()
	order.Seller = &abi.DoDSettleUser{Address: mock.Address()}
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := uo.DoGap(ctx, block)
	if err != nil || s != common.ContractDoDOrderState {
		t.Fatal()
	}

	order.ContractState = abi.DoDSettleContractStateConfirmed
	err = abi.DoDSettleUpdateOrder(ctx, order, param.InternalId)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err = uo.DoGap(ctx, block)
	if err != nil || s != common.ContractNoGap {
		t.Fatal()
	}
}

func TestDoDSettleChangeOrder_ProcessSend(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	co := new(DoDSettleChangeOrder)

	_, _, err := co.ProcessSend(ctx, block)
	if err != ErrToken {
		t.Fatal(err)
	}

	block.Token = cfg.GasToken()
	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(block.Address)
	am.Tokens[0] = mock.TokenMeta2(block.Address, cfg.GasToken())
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	cp := &abi.DoDSettleChangeOrderParam{
		Buyer:  &abi.DoDSettleUser{Address: mock.Address(), Name: "B1"},
		Seller: &abi.DoDSettleUser{Address: mock.Address(), Name: "S1"},
		Connections: []*abi.DoDSettleChangeConnectionParam{
			{
				ProductId: "p1",
				DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
					QuoteId:     "",
					ItemId:      "i1",
					QuoteItemId: "qi1",
					Bandwidth:   "100 Mbps",
					Price:       10,
					StartTime:   time.Now().Unix(),
				},
			},
		},
	}

	block.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	cp.Connections[0].QuoteId = "q1"
	block.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	cp.Buyer.Address = block.Address
	block.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	pid := &abi.DoDSettleProduct{Seller: cp.Seller.Address, ProductId: "p1"}
	otp := &abi.DoDSettleOrderToProduct{Seller: cp.Seller.Address, OrderId: "o1", OrderItemId: "oi1"}
	err = abi.DoDSettleSetProductStorageKeyByProductId(ctx, otp.Hash(), pid.Hash())
	if err != nil {
		t.Fatal(err)
	}

	err = abi.DoDSettleSetProductIdByStorageKey(ctx, otp.Hash(), "p1", cp.Seller.Address)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal(err)
	}

	conn := new(abi.DoDSettleConnectionParam)
	conn.BillingType = abi.DoDSettleBillingTypeDOD
	err = abi.DoDSettleUpdateConnectionRawParam(ctx, conn, otp.Hash())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, cp.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
	userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

	internalId := &abi.DoDSettleInternalIdWrap{InternalId: mock.Hash()}
	userInfo.InternalIds = append(userInfo.InternalIds, internalId)

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = co.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleChangeOrder_DoReceive(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	recv := mock.StateBlockWithoutWork()
	co := new(DoDSettleChangeOrder)

	_, err := co.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	param := new(abi.DoDSettleResponseParam)
	param.RequestHash = send.GetHash()
	param.Action = abi.DoDSettleResponseActionConfirm

	recv.Data, err = param.MarshalMsg(nil)
	if err != nil {
		t.Fatal()
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	order := new(abi.DoDSettleOrderInfo)
	order.Seller = &abi.DoDSettleUser{Address: recv.Address, Name: "s1"}
	order.Track = make([]*abi.DoDSettleOrderLifeTrack, 0)
	err = abi.DoDSettleUpdateOrder(ctx, order, send.Previous)
	if err != nil {
		t.Fatal(err)
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	param.Action = abi.DoDSettleResponseActionReject

	recv.Data, err = param.MarshalMsg(nil)
	if err != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(recv.Address)
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	am.Tokens[0] = mock.TokenMeta2(recv.Address, cfg.GasToken())
	err = l.UpdateAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = co.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleChangeOrder_GetTargetReceiver(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	co := new(DoDSettleChangeOrder)

	_, err := co.GetTargetReceiver(ctx, send)
	if err == nil {
		t.Fatal()
	}

	cp := &abi.DoDSettleChangeOrderParam{
		Buyer:  &abi.DoDSettleUser{Address: mock.Address(), Name: "B1"},
		Seller: &abi.DoDSettleUser{Address: mock.Address(), Name: "S1"},
	}

	send.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal()
	}

	tr, err := co.GetTargetReceiver(ctx, send)
	if err != nil || tr != cp.Seller.Address {
		t.Fatal()
	}
}

func TestDoDSettleTerminateOrder_ProcessSend(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	to := new(DoDSettleTerminateOrder)

	_, _, err := to.ProcessSend(ctx, block)
	if err != ErrToken {
		t.Fatal(err)
	}

	block.Token = cfg.GasToken()
	_, _, err = to.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(block.Address)
	am.Tokens[0] = mock.TokenMeta2(block.Address, cfg.GasToken())
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = to.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param := &abi.DoDSettleTerminateOrderParam{
		Buyer:  &abi.DoDSettleUser{Address: mock.Address(), Name: "B1"},
		Seller: &abi.DoDSettleUser{Address: mock.Address(), Name: "S1"},
		Connections: []*abi.DoDSettleChangeConnectionParam{
			{},
		},
	}

	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = to.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param.Connections[0].ItemId = "i1"
	param.Connections[0].ProductId = "p1"
	param.Connections[0].QuoteId = "q1"
	param.Connections[0].QuoteItemId = "qi1"
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	pid := &abi.DoDSettleProduct{Seller: param.Seller.Address, ProductId: "p1"}
	otp := &abi.DoDSettleOrderToProduct{Seller: param.Seller.Address, OrderId: "o1", OrderItemId: "oi1"}
	err = abi.DoDSettleSetProductStorageKeyByProductId(ctx, otp.Hash(), pid.Hash())
	if err != nil {
		t.Fatal(err)
	}

	err = abi.DoDSettleSetProductIdByStorageKey(ctx, otp.Hash(), "p1", param.Seller.Address)
	if err != nil {
		t.Fatal(err)
	}

	conn := new(abi.DoDSettleConnectionInfo)
	conn.ProductId = "p1"
	err = abi.DoDSettleUpdateConnection(ctx, conn, otp.Hash())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = to.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param.Buyer.Address = block.Address
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	cp := new(abi.DoDSettleConnectionParam)
	cp.BillingType = abi.DoDSettleBillingTypeDOD
	err = abi.DoDSettleUpdateConnectionRawParam(ctx, cp, otp.Hash())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = to.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, param.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
	userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

	internalId := &abi.DoDSettleInternalIdWrap{InternalId: mock.Hash()}
	userInfo.InternalIds = append(userInfo.InternalIds, internalId)

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = to.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleTerminateOrder_DoReceive(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	recv := mock.StateBlockWithoutWork()
	to := new(DoDSettleTerminateOrder)

	_, err := to.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	param := new(abi.DoDSettleResponseParam)
	param.RequestHash = send.GetHash()
	param.Action = abi.DoDSettleResponseActionConfirm

	recv.Data, err = param.MarshalMsg(nil)
	if err != nil {
		t.Fatal()
	}

	_, err = to.DoReceive(ctx, recv, send)
	if err == nil {
		t.Fatal()
	}

	order := new(abi.DoDSettleOrderInfo)
	order.Seller = &abi.DoDSettleUser{Address: recv.Address, Name: "s1"}
	order.Track = make([]*abi.DoDSettleOrderLifeTrack, 0)
	err = abi.DoDSettleUpdateOrder(ctx, order, send.Previous)
	if err != nil {
		t.Fatal(err)
	}

	_, err = to.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	param.Action = abi.DoDSettleResponseActionReject

	recv.Data, err = param.MarshalMsg(nil)
	if err != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(recv.Address)
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = to.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	am.Tokens[0] = mock.TokenMeta2(recv.Address, cfg.GasToken())
	err = l.UpdateAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, err = to.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettleTerminateOrder_GetTargetReceiver(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	send := mock.StateBlockWithoutWork()
	to := new(DoDSettleTerminateOrder)

	_, err := to.GetTargetReceiver(ctx, send)
	if err == nil {
		t.Fatal()
	}

	cp := &abi.DoDSettleTerminateOrderParam{
		Buyer:  &abi.DoDSettleUser{Address: mock.Address(), Name: "B1"},
		Seller: &abi.DoDSettleUser{Address: mock.Address(), Name: "S1"},
	}

	send.Data, err = cp.ToABI()
	if err != nil {
		t.Fatal()
	}

	tr, err := to.GetTargetReceiver(ctx, send)
	if err != nil || tr != cp.Seller.Address {
		t.Fatal()
	}
}

func TestDoDSettleUpdateProductInfo_ProcessSend(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	rr := new(DoDSettleUpdateProductInfo)

	_, _, err := rr.ProcessSend(ctx, block)
	if err != ErrToken {
		t.Fatal(err)
	}

	block.Token = cfg.GasToken()
	_, _, err = rr.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	am := mock.AccountMeta(block.Address)
	am.Tokens[0] = mock.TokenMeta2(block.Address, cfg.GasToken())
	err = l.AddAccountMeta(am, l.Cache().GetCache())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param := &abi.DoDSettleUpdateProductInfoParam{
		Address:     mock.Address(),
		OrderId:     "order1",
		ProductInfo: nil,
	}

	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param.ProductInfo = []*abi.DoDSettleProductInfo{{OrderItemId: "oi1", ProductId: "p1", Active: true}}
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	internalId := mock.Hash()
	order := abi.NewOrderInfo()
	order.OrderId = "order1"
	order.Seller = &abi.DoDSettleUser{Address: block.Address}
	order.OrderState = abi.DoDSettleOrderStateFail
	order.Connections = []*abi.DoDSettleConnectionParam{{}}
	order.Connections[0].OrderItemId = "oi1"
	order.Connections[0].ProductId = "p1"
	order.Connections[0].BillingType = abi.DoDSettleBillingTypePAYG
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	orderKey := &abi.DoDSettleOrder{
		Seller:  order.Seller.Address,
		OrderId: order.OrderId,
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableOrderIdMap)
	key = append(key, orderKey.Hash().Bytes()...)

	err = ctx.SetStorage(nil, key, internalId.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order.Seller = &abi.DoDSettleUser{Address: mock.Address()}
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order.Seller = &abi.DoDSettleUser{Address: block.Address}
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	order.OrderType = abi.DoDSettleOrderTypeCreate
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	pid := &abi.DoDSettleProduct{Seller: order.Seller.Address, ProductId: "p1"}
	otp := &abi.DoDSettleOrderToProduct{Seller: order.Seller.Address, OrderId: order.OrderId, OrderItemId: "oi1"}
	err = abi.DoDSettleSetProductStorageKeyByProductId(ctx, otp.Hash(), pid.Hash())
	if err != nil {
		t.Fatal(err)
	}

	err = abi.DoDSettleSetProductIdByStorageKey(ctx, otp.Hash(), "p1", order.Seller.Address)
	if err != nil {
		t.Fatal(err)
	}

	ci := &abi.DoDSettleConnectionInfo{
		Active: &abi.DoDSettleConnectionDynamicParam{
			BillingType: abi.DoDSettleBillingTypePAYG,
			BillingUnit: abi.DoDSettleBillingUnitSecond,
			StartTime:   1000,
		},
		Done:       make([]*abi.DoDSettleConnectionDynamicParam, 0),
		Disconnect: nil,
		Track:      make([]*abi.DoDSettleConnectionLifeTrack, 0),
	}
	err = abi.DoDSettleUpdateConnection(ctx, ci, otp.Hash())
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	internalId = mock.Hash()
	param.ProductInfo = []*abi.DoDSettleProductInfo{{OrderItemId: "oi1", ProductId: "p2", Active: true}}
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	order.Connections[0].ProductId = "p2"
	order.Connections[0].BillingType = abi.DoDSettleBillingTypePAYG
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	order.OrderType = abi.DoDSettleOrderTypeChange
	order.OrderId = "o3"
	order.Connections[0].ProductId = "p3"

	internalId = mock.Hash()
	param.ProductInfo = []*abi.DoDSettleProductInfo{{OrderItemId: "oi1", ProductId: "p3", Active: true}}
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal()
	}

	ph := abi.DoDSettleProduct{Seller: order.Seller.Address, ProductId: "p4"}

	ci = &abi.DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{},
		Active: &abi.DoDSettleConnectionDynamicParam{
			OrderId:     "o3",
			BillingType: abi.DoDSettleBillingTypePAYG,
			BillingUnit: abi.DoDSettleBillingUnitSecond,
			Price:       2,
			StartTime:   0,
			EndTime:     0,
		},
		Done: []*abi.DoDSettleConnectionDynamicParam{
			{
				OrderId:     "o1",
				BillingType: abi.DoDSettleBillingTypeDOD,
				Price:       2,
				StartTime:   40,
				EndTime:     50,
			},
			{
				OrderId:     "o2",
				BillingType: abi.DoDSettleBillingTypePAYG,
				BillingUnit: abi.DoDSettleBillingUnitSecond,
				Price:       2,
				StartTime:   10,
				EndTime:     0,
			},
		},
	}
	err = abi.DoDSettleUpdateConnection(ctx, ci, ph.Hash())
	if err != nil {
		t.Fatal()
	}

	internalId = mock.Hash()
	param.ProductInfo = []*abi.DoDSettleProductInfo{{OrderItemId: "oi1", ProductId: "p4", Active: true}}
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	order.Connections[0].ProductId = "p4"
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}

	ph = abi.DoDSettleProduct{Seller: order.Seller.Address, ProductId: "p5"}

	ci = &abi.DoDSettleConnectionInfo{
		DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{},
		Active: &abi.DoDSettleConnectionDynamicParam{
			OrderId:     "o3",
			BillingType: abi.DoDSettleBillingTypePAYG,
			BillingUnit: abi.DoDSettleBillingUnitSecond,
			Price:       2,
			StartTime:   0,
			EndTime:     0,
		},
		Done: []*abi.DoDSettleConnectionDynamicParam{
			{
				OrderId:     "o1",
				BillingType: abi.DoDSettleBillingTypeDOD,
				Price:       2,
				StartTime:   40,
				EndTime:     50,
			},
			{
				OrderId:     "o2",
				BillingType: abi.DoDSettleBillingTypePAYG,
				BillingUnit: abi.DoDSettleBillingUnitSecond,
				Price:       2,
				StartTime:   10,
				EndTime:     0,
			},
		},
	}
	err = abi.DoDSettleUpdateConnection(ctx, ci, ph.Hash())
	if err != nil {
		t.Fatal()
	}

	internalId = mock.Hash()
	param.ProductInfo = []*abi.DoDSettleProductInfo{{OrderItemId: "oi1", ProductId: "p5", Active: true}}
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	order.Connections[0].ProductId = "p5"
	order.OrderType = abi.DoDSettleOrderTypeTerminate
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.ProcessSend(ctx, block)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettleUpdateProductInfo_DoGap(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress)
	block := mock.StateBlockWithoutWork()
	rr := new(DoDSettleUpdateProductInfo)

	_, _, err := rr.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	param := &abi.DoDSettleUpdateProductInfoParam{
		Address: mock.Address(),
	}

	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	internalId := mock.Hash()
	param.OrderId = "order1"
	block.Data, err = param.ToABI()
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	orderKey := &abi.DoDSettleOrder{
		Seller:  block.Address,
		OrderId: param.OrderId,
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableOrderIdMap)
	key = append(key, orderKey.Hash().Bytes()...)

	err = ctx.SetStorage(nil, key, internalId.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	order := abi.NewOrderInfo()
	order.OrderState = abi.DoDSettleOrderStateFail
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = rr.DoGap(ctx, block)
	if err == nil {
		t.Fatal()
	}

	order.OrderState = abi.DoDSettleOrderStateNull
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err := rr.DoGap(ctx, block)
	if err != nil || s != common.ContractDoDOrderState {
		t.Fatal(err)
	}

	order.OrderState = abi.DoDSettleOrderStateSuccess
	err = abi.DoDSettleUpdateOrder(ctx, order, internalId)
	if err != nil {
		t.Fatal(err)
	}

	s, _, err = rr.DoGap(ctx, block)
	if err != nil || s != common.ContractNoGap {
		t.Fatal()
	}
}
