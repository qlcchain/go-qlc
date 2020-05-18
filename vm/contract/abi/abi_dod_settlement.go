package abi

import (
	"strings"
	"sync"
	"time"

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
	DoDSettlementLock   = new(DoDSettleLock)
)

const (
	DoDSettleLockHashSize = 1024
)

type DoDSettleLock struct {
	OrderInfoLock [DoDSettleLockHashSize]*sync.Mutex
}

func (l *DoDSettleLock) GetOrderInfoLock(id types.Hash) *sync.Mutex {
	index := (int(id[0])*256 + int(id[1])) % DoDSettleLockHashSize

	if l.OrderInfoLock[index] == nil {
		l.OrderInfoLock[index] = new(sync.Mutex)
	}

	return l.OrderInfoLock[index]
}

func DoDSettleBillingUnitRound(unit DoDSettleBillingUnit, s int64) int64 {
	now := time.Now()
	start := time.Unix(s, 0)
	var end time.Time

	switch unit {
	case DoDSettleBillingUnitYear:
		for {
			end = start.AddDate(1, 0, 0)
			if end.After(now) {
				break
			}
		}
		return end.Unix()
	case DoDSettleBillingUnitMonth:
		for {
			end = start.AddDate(0, 1, 0)
			if end.After(now) {
				break
			}
		}
		return end.Unix()
	case DoDSettleBillingUnitWeek:
		round := int64(60 * 60 * 24 * 7)
		return s + (now.Unix()-s+round-1)/round*round
	case DoDSettleBillingUnitDay:
		round := int64(60 * 60 * 24)
		return s + (now.Unix()-s+round-1)/round*round
	case DoDSettleBillingUnitHour:
		round := int64(60 * 60)
		return s + (now.Unix()-s+round-1)/round*round
	case DoDSettleBillingUnitMinute:
		round := int64(60)
		return s + (now.Unix()-s+round-1)/round*round
	case DoDSettleBillingUnitSecond:
		return start.Unix()
	default:
		return start.Unix()
	}
}

func DoDSettleCalcBillingUnit(unit DoDSettleBillingUnit, s, e int64) int {
	start := time.Unix(s, 0)
	end := time.Unix(e, 0)

	switch unit {
	case DoDSettleBillingUnitYear:
		return end.Year() - start.Year()
	case DoDSettleBillingUnitMonth:
		count := 0
		for {
			start = start.AddDate(0, 1, 0)
			count++
			if end.Sub(start) <= 0 {
				break
			}
		}
		return count
	case DoDSettleBillingUnitWeek:
		round := 60 * 60 * 24 * 7
		return int(e-s) / round
	case DoDSettleBillingUnitDay:
		round := 60 * 60 * 24
		return int(e-s) / round
	case DoDSettleBillingUnitHour:
		round := 60 * 60
		return int(e-s) / round
	case DoDSettleBillingUnitMinute:
		round := 60
		return int(e-s) / round
	case DoDSettleBillingUnitSecond:
		return int(e - s)
	default:
		return int(e - s)
	}
}

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

func DodSettleGenerateInvoiceByOrder(ctx *vmstore.VMContext, orderId string) {

}
