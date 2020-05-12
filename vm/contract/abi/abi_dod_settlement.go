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

func DoDSettleGetInternalIdByProductId(ctx *vmstore.VMContext, orderId string) (types.Hash, error) {
	orderIdHash := types.Sha256DHashData([]byte(orderId))

	var key []byte
	key = append(key, DoDSettleDBTableIdMap)
	key = append(key, orderIdHash.Bytes()...)
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

func DoDSettleGetOrderInfoByOrderId(ctx *vmstore.VMContext, orderId string) (*DoDSettleOrderInfo, error) {
	internalId, err := DoDSettleGetInternalIdByProductId(ctx, orderId)
	if err != nil {
		return nil, err
	}

	return DoDSettleGetOrderInfoByInternalId(ctx, internalId)
}

func DoDSettleGetConnectionInfoByProductId(ctx *vmstore.VMContext, productId types.Hash) (*DoDSettleConnectionInfo, error) {
	var key []byte
	key = append(key, DoDSettleDBTableProduct)
	key = append(key, productId.Bytes()...)
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
