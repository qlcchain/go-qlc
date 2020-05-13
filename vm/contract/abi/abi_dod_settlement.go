package abi

import (
	"strings"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

const (
	JsonDoDSettlement = `[
		{"type":"function","name":"DoDSettleCreateOrder","inputs":[
			{"name":"buyerAddress","type":"address"},
			{"name":"buyerName","type":"string"},
			{"name":"sellerAddress","type":"address"},
			{"name":"sellerName","type":"string"},
			{"name":"connectionName","type":"string"},
			{"name":"srcCompanyName","type":"string"},
			{"name":"srcRegion","type":"string"},
			{"name":"srcCity","type":"string"},
			{"name":"srcDataCenter","type":"string"},
			{"name":"srcPort","type":"string"},
			{"name":"dstCompanyName","type":"string"},
			{"name":"dstRegion","type":"string"},
			{"name":"dstCity","type":"string"},
			{"name":"dstDataCenter","type":"string"},
			{"name":"dstPort","type":"string"},
			{"name":"paymentType","type":"int64"},
			{"name":"billingType","type":"int64"},
			{"name":"currency","type":"string"},
			{"name":"bandwidth","type":"string"},
			{"name":"billingUnit","type":"int64"},
			{"name":"price","type":"string"},
			{"name":"startTime","type":"uint64"},
			{"name":"endTime","type":"uint64"},
			{"name":"fee","type":"string"},
			{"name":"serviceClass","type":"int64"}
		]},
		{"type":"function","name":"DoDSettleUpdateOrderInfo","inputs":[
			{"name":"buyer","type":"address"},
			{"name":"internalId","type":"hash"},
			{"name":"orderId","type":"string"},
			{"name":"productId","type":"string[]"},
			{"name":"operation","type":"string"},
			{"name":"failReason","type":"string"}
		]},
		{"type":"function","name":"DoDSettleChangeOrder","inputs":[
			{"name":"buyerAddress","type":"address"},
			{"name":"buyerName","type":"string"},
			{"name":"sellerAddress","type":"address"},
			{"name":"sellerName","type":"string"},
			{"name":"productId","type":"string"},
			{"name":"connectionName","type":"string"},
			{"name":"paymentType","type":"int64"},
			{"name":"billingType","type":"int64"},
			{"name":"currency","type":"string"},
			{"name":"bandwidth","type":"string"},
			{"name":"billingUnit","type":"int64"},
			{"name":"price","type":"string"},
			{"name":"startTime","type":"uint64"},
			{"name":"endTime","type":"uint64"},
			{"name":"fee","type":"string"},
			{"name":"serviceClass","type":"int64"}
		]},
		{"type":"function","name":"DoDSettleTerminateOrder","inputs":[
			{"name":"buyer","type":"address"},
			{"name":"orderId","type":"string"}
		]},
		{"type":"function","name":"DoDSettleResourceReady","inputs":[
			{"name":"seller","type":"address"},
			{"name":"orderId","type":"string"},
			{"name":"productId","type":"string"}
		]}
	]`

	MethodNameDoDSettleCreateOrder     = "DoDSettleCreateOrder"
	MethodNameDoDSettleUpdateOrderInfo = "DoDSettleUpdateOrderInfo"
	MethodNameDoDSettleChangeOrder     = "DoDSettleChangeOrder"
	MethodNameDoDSettleTerminateOrder  = "DoDSettleTerminateOrder"
	MethodNameDoDSettleResourceReady   = "DoDSettleResourceReady"
)

var (
	DoDSettlementABI, _ = abi.JSONToABIContract(strings.NewReader(JsonDoDSettlement))
)

func DoDSettleGetOrderInfoByInternalId(ctx *vmstore.VMContext, id types.Hash) (*DoDSettleOrderInfo, error) {
	var key []byte
	key = append(key, DoDSettleDBTableOrder)
	key = append(key, id.Bytes()...)
	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	oi := new(DoDSettleOrderInfo)
	_, err = oi.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return oi, nil
}

func DoDSettleGetInternalIdByOrderId(ctx *vmstore.VMContext, seller types.Address, orderId string) (types.Hash, error) {
	orderKey := &DoDSettleOrder{Seller: seller, OrderId: orderId}

	var key []byte
	key = append(key, DoDSettleDBTableOrderIdMap)
	key = append(key, orderKey.Hash().Bytes()...)
	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return types.ZeroHash, err
	}

	hash, err := types.BytesToHash(data)
	if err != nil {
		return types.ZeroHash, err
	}

	return hash, nil
}

func DoDSettleGetOrderInfoByOrderId(ctx *vmstore.VMContext, seller types.Address, orderId string) (*DoDSettleOrderInfo, error) {
	internalId, err := DoDSettleGetInternalIdByOrderId(ctx, seller, orderId)
	if err != nil {
		return nil, err
	}

	return DoDSettleGetOrderInfoByInternalId(ctx, internalId)
}

func DoDSettleGetConnectionInfoByProductId(ctx *vmstore.VMContext, seller types.Address, productId string) (*DoDSettleConnectionInfo, error) {
	productKey := &DoDSettleProduct{Seller: seller, ProductId: productId}

	var key []byte
	key = append(key, DoDSettleDBTableProduct)
	key = append(key, productKey.Hash().Bytes()...)
	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	conn := new(DoDSettleConnectionInfo)
	_, err = conn.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func DoDSettleGetConnectionInfoByProductHash(ctx *vmstore.VMContext, hash types.Hash) (*DoDSettleConnectionInfo, error) {
	var key []byte
	key = append(key, DoDSettleDBTableProduct)
	key = append(key, hash.Bytes()...)
	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	conn := new(DoDSettleConnectionInfo)
	_, err = conn.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func DoDSettleUpdateOrder(ctx *vmstore.VMContext, order *DoDSettleOrderInfo, id types.Hash) error {
	data, err := order.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, DoDSettleDBTableOrder)
	key = append(key, id.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}

func DoDSettleInheritParam(src, dst *DoDSettleConnectionDynamicParam) {
	if len(dst.ConnectionName) == 0 {
		dst.ConnectionName = src.ConnectionName
	}

	if len(dst.Currency) == 0 {
		dst.Currency = src.Currency
	}

	if dst.BillingType == 0 {
		dst.BillingType = src.BillingType
	}

	if dst.BillingUnit == 0 {
		dst.BillingUnit = src.BillingUnit
	}

	if len(dst.Bandwidth) == 0 {
		dst.Bandwidth = src.Bandwidth
	}

	if dst.ServiceClass == 0 {
		dst.ServiceClass = src.ServiceClass
	}

	if dst.PaymentType == 0 {
		dst.PaymentType = src.PaymentType
	}
}

func DoDSettleGetInternalIdListByAddress(ctx *vmstore.VMContext, address types.Address) ([]types.Hash, error) {
	var key []byte
	key = append(key, DoDSettleDBTableUser)
	key = append(key, address.Bytes()...)

	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	userInfo := new(DoDSettleUserInfos)
	_, err = userInfo.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	hs := make([]types.Hash, 0)
	for _, i := range userInfo.InternalIds {
		hs = append(hs, i.InternalId)
	}

	return hs, nil
}

func DoDSettleGetProductIdListByAddress(ctx *vmstore.VMContext, address types.Address) ([]*DoDSettleProduct, error) {
	var key []byte
	key = append(key, DoDSettleDBTableUser)
	key = append(key, address.Bytes()...)

	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	userInfo := new(DoDSettleUserInfos)
	_, err = userInfo.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return userInfo.ProductIds, nil
}

func DoDSettleGetOrderIdListByAddress(ctx *vmstore.VMContext, address types.Address) ([]*DoDSettleOrder, error) {
	var key []byte
	key = append(key, DoDSettleDBTableUser)
	key = append(key, address.Bytes()...)

	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	userInfo := new(DoDSettleUserInfos)
	_, err = userInfo.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return userInfo.OrderIds, nil
}
