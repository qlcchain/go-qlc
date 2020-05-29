package api

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"testing"
)

func DoDSettleAPITestInit(t *testing.T) (*DoDSettlementAPI, func()) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}

	return NewDoDSettlementAPI(cfgFile, l), clear
}

func TestDoDSettlementAPI_GetCreateOrderBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleCreateOrderParam)
	param.Buyer = &abi.DoDSettleUser{Address: mock.Address()}

	_, err := ds.GetCreateOrderBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetCreateOrderBlock(param)
}

func TestDoDSettlementAPI_GetCreateOrderRewardBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleResponseParam)

	_, err := ds.GetCreateOrderRewardBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetCreateOrderRewardBlock(param)
}

func TestDoDSettlementAPI_GetUpdateOrderInfoBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleUpdateOrderInfoParam)

	_, err := ds.GetUpdateOrderInfoBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetUpdateOrderInfoBlock(param)
}

func TestDoDSettlementAPI_GetUpdateOrderInfoRewardBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleResponseParam)

	_, err := ds.GetUpdateOrderInfoRewardBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetUpdateOrderInfoRewardBlock(param)
}

func TestDoDSettlementAPI_GetChangeOrderBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleChangeOrderParam)
	param.Buyer = &abi.DoDSettleUser{Address: mock.Address()}

	_, err := ds.GetChangeOrderBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetChangeOrderBlock(param)
}

func TestDoDSettlementAPI_GetChangeOrderRewardBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleResponseParam)

	_, err := ds.GetChangeOrderRewardBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetChangeOrderRewardBlock(param)
}

func TestDoDSettlementAPI_GetTerminateOrderBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleTerminateOrderParam)
	param.Buyer = &abi.DoDSettleUser{Address: mock.Address()}

	_, err := ds.GetTerminateOrderBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetTerminateOrderBlock(param)
}

func TestDoDSettlementAPI_GetTerminateOrderRewardBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleResponseParam)

	_, err := ds.GetTerminateOrderRewardBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetTerminateOrderRewardBlock(param)
}

func TestDoDSettlementAPI_GetResourceReadyBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleResourceReadyParam)

	_, err := ds.GetResourceReadyBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetResourceReadyBlock(param)
}

func TestDoDSettlementAPI_GetResourceReadyRewardBlock(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	param := new(DoDSettleResponseParam)

	_, err := ds.GetResourceReadyRewardBlock(nil)
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetResourceReadyRewardBlock(param)
}

func TestDoDSettlementAPI_GetOrderInfoBySellerAndOrderId(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GetOrderInfoBySellerAndOrderId(mock.Address(), "order001")
}

func TestDoDSettlementAPI_GetOrderInfoByInternalId(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	_, err := ds.GetOrderInfoByInternalId("123")
	if err == nil {
		t.Fatal()
	}

	_, _ = ds.GetOrderInfoByInternalId("63be22932dd23059ad3706e347d0b8343de752d9bff9d12f5132102d0bd13b9b")
}

func TestDoDSettlementAPI_GetConnectionInfoBySellerAndProductId(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GetConnectionInfoBySellerAndProductId(mock.Address(), "product001")
}

func TestDoDSettlementAPI_GetPendingRequest(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	seller := mock.Address()
	_, err := ds.GetPendingRequest(seller)
	if err != nil {
		t.Fatal()
	}

	block := mock.StateBlockWithoutWork()
	block.Type = types.ContractSend
	block.Link = contractaddress.DoDSettlementAddress.ToHash()

	pk := &types.PendingKey{
		Address: seller,
		Hash:    block.GetHash(),
	}
	pi := &types.PendingInfo{
		Source: block.Address,
		Amount: types.NewBalance(0),
		Type:   cfg.GasToken(),
	}

	err = ds.l.AddPending(pk, pi, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	if err := ds.l.Flush(); err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPendingRequest(seller)
	if err == nil {
		t.Fatal()
	}

	err = ds.l.AddStateBlock(block)
	if err != nil {
		t.Fatal()
	}

	_, err = ds.GetPendingRequest(seller)
	if err == nil {
		t.Fatal()
	}

	param := new(abi.DoDSettleCreateOrderParam)
	block.Data, _ = param.ToABI()
	err = ds.l.AddStateBlock(block)
	if err != nil {
		t.Fatal()
	}

	err = ds.l.DeletePending(pk, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	pk.Hash = block.GetHash()
	err = ds.l.AddPending(pk, pi, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	if err := ds.l.Flush(); err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPendingRequest(seller)
	if err == nil {
		t.Fatal()
	}

	order := abi.NewOrderInfo()
	err = abi.DoDSettleUpdateOrder(ds.ctx, order, block.Previous)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPendingRequest(seller)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoDSettlementAPI_GetPendingResourceCheck(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	seller := mock.Address()
	_, err := ds.GetPendingResourceCheck(seller)
	if err != nil {
		t.Fatal()
	}

	block := mock.StateBlockWithoutWork()
	block.Link = contractaddress.DoDSettlementAddress.ToHash()
	block.Type = types.ContractSend

	pk := &types.PendingKey{
		Address: seller,
		Hash:    block.GetHash(),
	}
	pi := &types.PendingInfo{
		Source: block.Address,
		Amount: types.NewBalance(0),
		Type:   cfg.GasToken(),
	}

	err = ds.l.AddPending(pk, pi, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	if err := ds.l.Flush(); err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPendingResourceCheck(seller)
	if err == nil {
		t.Fatal()
	}

	err = ds.l.AddStateBlock(block)
	if err != nil {
		t.Fatal()
	}

	_, err = ds.GetPendingResourceCheck(seller)
	if err == nil {
		t.Fatal()
	}

	param := new(abi.DoDSettleUpdateOrderInfoParam)
	param.Status = abi.DoDSettleOrderStateFail
	block.Data, _ = param.ToABI()
	err = ds.l.AddStateBlock(block)
	if err != nil {
		t.Fatal()
	}

	err = ds.l.DeletePending(pk, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	pk.Hash = block.GetHash()
	err = ds.l.AddPending(pk, pi, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	if err := ds.l.Flush(); err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPendingResourceCheck(seller)
	if err != nil {
		t.Fatal()
	}

	param.Status = abi.DoDSettleOrderStateSuccess
	param.ProductIds = []*abi.DoDSettleProductItem{{ProductId: "product001", BuyerProductId: "bp1"}}
	block.Data, _ = param.ToABI()
	err = ds.l.AddStateBlock(block)
	if err != nil {
		t.Fatal()
	}

	err = ds.l.DeletePending(pk, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	pk.Hash = block.GetHash()
	err = ds.l.AddPending(pk, pi, ds.l.Cache().GetCache())
	if err != nil {
		t.Fatal()
	}

	if err := ds.l.Flush(); err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPendingResourceCheck(seller)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettlementAPI_GetPlacingOrder(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	buyer := mock.Address()
	seller := mock.Address()
	id := mock.Hash()

	_, err := ds.GetPlacingOrder(buyer, seller)
	if err == nil {
		t.Fatal()
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, buyer.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
	userInfo.ProductIds = make([]*abi.DoDSettleProduct, 0)
	userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

	internalId := &abi.DoDSettleInternalIdWrap{InternalId: id}
	userInfo.InternalIds = append(userInfo.InternalIds, internalId)

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ds.ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPlacingOrder(buyer, seller)
	if err != nil {
		t.Fatal()
	}

	order := abi.NewOrderInfo()
	err = abi.DoDSettleUpdateOrder(ds.ctx, order, id)
	if err != nil {
		t.Fatal(err)
	}

	order.Seller = &abi.DoDSettleUser{Address: seller}
	order.ContractState = abi.DoDSettleContractStateConfirmed
	order.OrderState = abi.DoDSettleOrderStateNull
	err = abi.DoDSettleUpdateOrder(ds.ctx, order, id)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ds.GetPlacingOrder(buyer, seller)
	if err != nil {
		t.Fatal()
	}
}

func TestDoDSettlementAPI_GetProductIdListByAddress(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GetProductIdListByAddress(mock.Address())
}

func TestDoDSettlementAPI_GetOrderIdListByAddress(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GetOrderIdListByAddress(mock.Address())
}

func TestDoDSettlementAPI_GetProductIdListByAddressAndSeller(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	seller := mock.Address()
	buyer := mock.Address()
	_, err := ds.GetProductIdListByAddressAndSeller(buyer, seller)
	if err == nil {
		t.Fatal()
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, buyer.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.ProductIds = []*abi.DoDSettleProduct{{Seller: seller, ProductId: "p1"}, {Seller: mock.Address(), ProductId: "p2"}}

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ds.ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	pd, err := ds.GetProductIdListByAddressAndSeller(buyer, seller)
	if err != nil || len(pd) == 0 || pd[0].ProductId != "p1" {
		t.Fatal()
	}
}

func TestDoDSettlementAPI_GetOrderIdListByAddressAndSeller(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()

	seller := mock.Address()
	buyer := mock.Address()
	_, err := ds.GetOrderIdListByAddressAndSeller(buyer, seller)
	if err == nil {
		t.Fatal()
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, buyer.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)
	userInfo.OrderIds = []*abi.DoDSettleOrder{{Seller: seller, OrderId: "o1"}, {Seller: mock.Address(), OrderId: "o2"}}

	data, err := userInfo.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = ds.ctx.SetStorage(nil, key, data)
	if err != nil {
		t.Fatal(err)
	}

	or, err := ds.GetOrderIdListByAddressAndSeller(buyer, seller)
	if err != nil || len(or) == 0 || or[0].OrderId != "o1" {
		t.Fatal()
	}
}

func TestDoDSettlementAPI_GenerateInvoiceByBuyer(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GenerateInvoiceByBuyer(mock.Address(), mock.Address(), 100, 1000, true, true)
}

func TestDoDSettlementAPI_GenerateInvoiceByOrderId(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GenerateInvoiceByOrderId(mock.Address(), "order1", 100, 1000, true, true)
}

func TestDoDSettlementAPI_GenerateInvoiceByProductId(t *testing.T) {
	ds, clear := DoDSettleAPITestInit(t)
	defer clear()
	_, _ = ds.GenerateInvoiceByProductId(mock.Address(), "product1", 100, 1000, true, true)
}
