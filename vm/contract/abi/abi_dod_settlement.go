package abi

import (
	"fmt"
	"strings"
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
)

func DoDSettleBillingUnitRound(unit DoDSettleBillingUnit, s, t int64) int64 {
	to := time.Unix(t, 0)
	start := time.Unix(s, 0)
	var end time.Time

	if s == t {
		return s
	}

	switch unit {
	case DoDSettleBillingUnitYear:
		for {
			end = start.AddDate(1, 0, 0)
			if end.After(to) || end.Equal(to) {
				break
			}
		}
		return end.Unix()
	case DoDSettleBillingUnitMonth:
		for {
			end = start.AddDate(0, 1, 0)
			if end.After(to) || end.Equal(to) {
				break
			}
		}
		return end.Unix()
	case DoDSettleBillingUnitWeek:
		round := int64(60 * 60 * 24 * 7)
		return s + (t-s+round-1)/round*round
	case DoDSettleBillingUnitDay:
		round := int64(60 * 60 * 24)
		return s + (t-s+round-1)/round*round
	case DoDSettleBillingUnitHour:
		round := int64(60 * 60)
		return s + (t-s+round-1)/round*round
	case DoDSettleBillingUnitMinute:
		round := int64(60)
		return s + (t-s+round-1)/round*round
	case DoDSettleBillingUnitSecond:
		return t
	default:
		return t
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

func DoDSettleUpdateConnection(ctx *vmstore.VMContext, conn *DoDSettleConnectionInfo, id types.Hash) error {
	data, err := conn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, DoDSettleDBTableProduct)
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

func DodSettleGetOrderInvoice(ctx *vmstore.VMContext, seller types.Address, order *DoDSettleOrderInfo, start, end int64) (*DoDSettleInvoiceOrderDetail, error) {
	invoiceOrder := new(DoDSettleInvoiceOrderDetail)
	invoiceOrder.OrderId = order.OrderId
	invoiceOrder.Connections = make([]*DoDSettleInvoiceConnDetail, 0)

	for _, c := range order.Connections {
		conn, _ := DoDSettleGetConnectionInfoByProductId(ctx, seller, c.ProductId)
		if conn == nil {
			continue
		}

		ic := &DoDSettleInvoiceConnDetail{
			DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
				ProductId:      conn.ProductId,
				SrcCompanyName: conn.SrcCompanyName,
				SrcRegion:      conn.SrcRegion,
				SrcCity:        conn.SrcCity,
				SrcDataCenter:  conn.SrcDataCenter,
				SrcPort:        conn.SrcPort,
				DstCompanyName: conn.DstCompanyName,
				DstRegion:      conn.DstRegion,
				DstCity:        conn.DstCity,
				DstDataCenter:  conn.DstDataCenter,
				DstPort:        conn.DstPort,
			},
			Usage: make([]*DoDSettleInvoiceConnDynamic, 0),
		}

		if conn.Active != nil && conn.Active.OrderId == order.OrderId && end > conn.Active.StartTime {
			dc := &DoDSettleInvoiceConnDynamic{
				DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
					ConnectionName: conn.Active.ConnectionName,
					PaymentType:    conn.Active.PaymentType,
					BillingType:    conn.Active.BillingType,
					Currency:       conn.Active.Currency,
					ServiceClass:   conn.Active.ServiceClass,
					Bandwidth:      conn.Active.Bandwidth,
					BillingUnit:    conn.Active.BillingUnit,
					Price:          conn.Active.Price,
					StartTime:      conn.Active.StartTime,
					EndTime:        conn.Active.EndTime,
				},
			}

			if conn.Active.BillingType == DoDSettleBillingTypeDOD {
				dc.Amount = conn.Active.Price
			} else {
				if start > conn.Active.StartTime {
					dc.InvoiceStartTime = DoDSettleBillingUnitRound(conn.Active.BillingUnit, conn.Active.StartTime, start)
				} else {
					dc.InvoiceStartTime = conn.Active.StartTime
				}

				dc.InvoiceEndTime = DoDSettleBillingUnitRound(conn.Active.BillingUnit, conn.Active.StartTime, end)
				dc.InvoiceUnitCount = DoDSettleCalcBillingUnit(conn.Active.BillingUnit, dc.InvoiceStartTime, dc.InvoiceEndTime)
				dc.Amount = conn.Active.Price * float64(dc.InvoiceUnitCount)
			}

			ic.ConnectionAmount += dc.Amount
			ic.Usage = append(ic.Usage, dc)
		}

		for _, done := range conn.Done {
			if done.OrderId == order.OrderId && ((start >= done.StartTime && start <= done.EndTime) ||
				(end > done.StartTime && end < done.EndTime) ||
				(start < done.StartTime && end > done.EndTime)) {
				dc := &DoDSettleInvoiceConnDynamic{
					DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
						ConnectionName: done.ConnectionName,
						PaymentType:    done.PaymentType,
						BillingType:    done.BillingType,
						Currency:       done.Currency,
						ServiceClass:   done.ServiceClass,
						Bandwidth:      done.Bandwidth,
						BillingUnit:    done.BillingUnit,
						Price:          done.Price,
						StartTime:      done.StartTime,
						EndTime:        done.EndTime,
					},
				}

				if done.BillingType == DoDSettleBillingTypeDOD {
					dc.Amount = done.Price
				} else {
					if start >= done.StartTime {
						dc.InvoiceStartTime = DoDSettleBillingUnitRound(done.BillingUnit, done.StartTime, start)
					} else {
						dc.InvoiceStartTime = done.StartTime
					}

					if end < done.EndTime {
						dc.InvoiceEndTime = DoDSettleBillingUnitRound(done.BillingUnit, done.StartTime, end)
					} else {
						dc.InvoiceEndTime = done.EndTime
					}

					dc.InvoiceUnitCount = DoDSettleCalcBillingUnit(done.BillingUnit, dc.InvoiceStartTime, dc.InvoiceEndTime)
					dc.Amount = done.Price * float64(dc.InvoiceUnitCount)
				}

				ic.ConnectionAmount += dc.Amount
				ic.Usage = append(ic.Usage, dc)
				break
			}
		}

		invoiceOrder.OrderAmount += ic.ConnectionAmount
		invoiceOrder.ConnectionCount++
		invoiceOrder.Connections = append(invoiceOrder.Connections, ic)
	}

	return invoiceOrder, nil
}

func DodSettleGetProductInvoice(ctx *vmstore.VMContext, seller types.Address, conn *DoDSettleConnectionInfo, start, end int64) (*DoDSettleInvoiceConnDetail, error) {
	invoiceProduct := &DoDSettleInvoiceConnDetail{
		DoDSettleConnectionStaticParam: DoDSettleConnectionStaticParam{
			ProductId:      conn.ProductId,
			SrcCompanyName: conn.SrcCompanyName,
			SrcRegion:      conn.SrcRegion,
			SrcCity:        conn.SrcCity,
			SrcDataCenter:  conn.SrcDataCenter,
			SrcPort:        conn.SrcPort,
			DstCompanyName: conn.DstCompanyName,
			DstRegion:      conn.DstRegion,
			DstCity:        conn.DstCity,
			DstDataCenter:  conn.DstDataCenter,
			DstPort:        conn.DstPort,
		},
		Usage: make([]*DoDSettleInvoiceConnDynamic, 0),
	}

	if conn.Active != nil && end > conn.Active.StartTime {
		dc := &DoDSettleInvoiceConnDynamic{
			DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
				OrderId:        conn.Active.OrderId,
				ConnectionName: conn.Active.ConnectionName,
				PaymentType:    conn.Active.PaymentType,
				BillingType:    conn.Active.BillingType,
				Currency:       conn.Active.Currency,
				ServiceClass:   conn.Active.ServiceClass,
				Bandwidth:      conn.Active.Bandwidth,
				BillingUnit:    conn.Active.BillingUnit,
				Price:          conn.Active.Price,
				StartTime:      conn.Active.StartTime,
				EndTime:        conn.Active.EndTime,
			},
		}

		if conn.Active.BillingType == DoDSettleBillingTypeDOD {
			dc.Amount = conn.Active.Price
		} else {
			if start > conn.Active.StartTime {
				dc.InvoiceStartTime = DoDSettleBillingUnitRound(conn.Active.BillingUnit, conn.Active.StartTime, start)
			} else {
				dc.InvoiceStartTime = conn.Active.StartTime
			}

			dc.InvoiceEndTime = DoDSettleBillingUnitRound(conn.Active.BillingUnit, conn.Active.StartTime, end)
			dc.InvoiceUnitCount = DoDSettleCalcBillingUnit(conn.Active.BillingUnit, dc.InvoiceStartTime, dc.InvoiceEndTime)
			dc.Amount = conn.Active.Price * float64(dc.InvoiceUnitCount)
		}

		invoiceProduct.ConnectionAmount += dc.Amount
		invoiceProduct.Usage = append(invoiceProduct.Usage, dc)
	}

	for _, done := range conn.Done {
		if (start >= done.StartTime && start <= done.EndTime) ||
			(end > done.StartTime && end < done.EndTime) ||
			(start < done.StartTime && end > done.EndTime) {
			dc := &DoDSettleInvoiceConnDynamic{
				DoDSettleConnectionDynamicParam: DoDSettleConnectionDynamicParam{
					OrderId:        done.OrderId,
					ConnectionName: done.ConnectionName,
					PaymentType:    done.PaymentType,
					BillingType:    done.BillingType,
					Currency:       done.Currency,
					ServiceClass:   done.ServiceClass,
					Bandwidth:      done.Bandwidth,
					BillingUnit:    done.BillingUnit,
					Price:          done.Price,
					StartTime:      done.StartTime,
					EndTime:        done.EndTime,
				},
			}

			if done.BillingType == DoDSettleBillingTypeDOD {
				dc.Amount = done.Price
			} else {
				if start >= done.StartTime {
					dc.InvoiceStartTime = DoDSettleBillingUnitRound(done.BillingUnit, done.StartTime, start)
				} else {
					dc.InvoiceStartTime = done.StartTime
				}

				if end < done.EndTime {
					dc.InvoiceEndTime = DoDSettleBillingUnitRound(done.BillingUnit, done.StartTime, end)
				} else {
					dc.InvoiceEndTime = done.EndTime
				}

				dc.InvoiceUnitCount = DoDSettleCalcBillingUnit(done.BillingUnit, dc.InvoiceStartTime, dc.InvoiceEndTime)
				dc.Amount = done.Price * float64(dc.InvoiceUnitCount)
			}

			invoiceProduct.ConnectionAmount += dc.Amount
			invoiceProduct.Usage = append(invoiceProduct.Usage, dc)
		}
	}

	return invoiceProduct, nil
}

func DodSettleGenerateInvoiceByOrder(ctx *vmstore.VMContext, seller types.Address, orderId string, start, end int64) (*DoDSettleOrderInvoice, error) {
	invoice := new(DoDSettleOrderInvoice)

	now := time.Now().Unix()
	if start > end || start > now || end > now {
		return nil, fmt.Errorf("invalid start or end time")
	}

	order, err := DoDSettleGetOrderInfoByOrderId(ctx, seller, orderId)
	if err != nil {
		return nil, err
	}

	invoiceOrder, err := DodSettleGetOrderInvoice(ctx, seller, order, start, end)
	if err != nil {
		return nil, err
	}

	invoice.StartTime = start
	invoice.EndTime = end
	invoice.Currency = order.Connections[0].Currency
	invoice.Buyer = order.Buyer
	invoice.Seller = order.Seller
	invoice.TotalConnectionCount = invoiceOrder.ConnectionCount
	invoice.TotalAmount = invoiceOrder.OrderAmount
	invoice.Order = invoiceOrder

	return invoice, nil
}

func DodSettleGenerateInvoiceByProduct(ctx *vmstore.VMContext, seller types.Address, productId string, start, end int64) (*DoDSettleProductInvoice, error) {
	invoice := new(DoDSettleProductInvoice)

	now := time.Now().Unix()
	if start > end || start > now || end > now {
		return nil, fmt.Errorf("invalid start or end time")
	}

	conn, err := DoDSettleGetConnectionInfoByProductId(ctx, seller, productId)
	if err != nil {
		return nil, err
	}

	order, err := DoDSettleGetOrderInfoByOrderId(ctx, seller, conn.Track[0].OrderId)
	if err != nil {
		return nil, err
	}

	productOrder, err := DodSettleGetProductInvoice(ctx, seller, conn, start, end)
	if err != nil {
		return nil, err
	}

	invoice.StartTime = start
	invoice.EndTime = end
	invoice.Currency = order.Connections[0].Currency
	invoice.Buyer = order.Buyer
	invoice.Seller = order.Seller
	invoice.TotalAmount = productOrder.ConnectionAmount
	invoice.Connection = productOrder

	return invoice, nil
}

func DodSettleGenerateInvoiceByBuyer(ctx *vmstore.VMContext, seller, buyer types.Address, start, end int64) (*DoDSettleBuyerInvoice, error) {
	invoice := new(DoDSettleBuyerInvoice)

	now := time.Now().Unix()
	if start > end || start > now || end > now {
		return nil, fmt.Errorf("invalid start or end time")
	}

	invoice.StartTime = start
	invoice.EndTime = end
	invoice.Orders = make([]*DoDSettleInvoiceOrderDetail, 0)

	orders, err := DoDSettleGetOrderIdListByAddress(ctx, buyer)
	if err != nil {
		return nil, err
	}

	for _, o := range orders {
		order, err := DoDSettleGetOrderInfoByOrderId(ctx, seller, o.OrderId)
		if err != nil {
			return nil, err
		}

		if order.OrderType == DoDSettleOrderTypeTerminate {
			continue
		}

		if invoice.Buyer == nil {
			invoice.Currency = order.Connections[0].Currency
			invoice.Buyer = order.Buyer
			invoice.Seller = order.Seller
		}

		invoiceOrder, err := DodSettleGetOrderInvoice(ctx, seller, order, start, end)
		if err != nil {
			return nil, err
		}

		invoice.OrderCount++
		invoice.TotalConnectionCount += invoiceOrder.ConnectionCount
		invoice.TotalAmount += invoiceOrder.OrderAmount
		invoice.Orders = append(invoice.Orders, invoiceOrder)
	}

	return invoice, nil
}

func DodSettleSetSellerConnectionActive(ctx *vmstore.VMContext, active *DoDSettleConnectionActive, id types.Hash) error {
	data, err := active.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, DoDSettleDBTableSellerConnActive)
	key = append(key, id.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}

func DodSettleGetSellerConnectionActive(ctx *vmstore.VMContext, id types.Hash) (*DoDSettleConnectionActive, error) {
	var key []byte
	key = append(key, DoDSettleDBTableSellerConnActive)
	key = append(key, id.Bytes()...)

	data, err := ctx.GetStorage(nil, key)
	if err != nil {
		return nil, err
	}

	act := new(DoDSettleConnectionActive)
	_, err = act.UnmarshalMsg(data)
	if err != nil {
		return nil, err
	}

	return act, nil
}
