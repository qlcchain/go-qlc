package abi

import (
	"fmt"

	"github.com/qlcchain/go-qlc/vm/vmstore"

	"github.com/qlcchain/go-qlc/common/types"
)

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
request
confirmed
rejected
)
*/
type DoDSettleContractState int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
success
complete
fail
)
*/
type DoDSettleOrderState int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
invoice
stableCoin
)
*/
type DoDSettlePaymentType int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
PAYG
DOD
)
*/
type DoDSettleBillingType int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
year
month
week
day
hour
minute
second
)
*/
type DoDSettleBillingUnit int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
gold
silver
bronze
)
*/
type DoDSettleServiceClass int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
confirm
reject
)
*/
type DoDSettleResponseAction int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
null
create
change
terminate
)
*/
type DoDSettleOrderType int

const (
	DoDSettleDBTableOrder uint8 = iota
	DoDSettleDBTableProduct
	DoDSettleDBTableOrderIdMap
	DoDSettleDBTableUser
	DoDSettleDBTableSellerConnActive
	DoDSettleDBTableBuyerConnActive
)

//go:generate msgp
type DoDSettleUser struct {
	Address types.Address `json:"address" msg:"a,extension"`
	Name    string        `json:"name" msg:"n"`
}

type DoDSettleInternalIdWrap struct {
	InternalId types.Hash `json:"id" msg:"i,extension"`
}

type DoDSettleUserInfos struct {
	InternalIds []*DoDSettleInternalIdWrap `json:"internalIds,omitempty" msg:"i"`
	ProductIds  []*DoDSettleProduct        `json:"productIds,omitempty" msg:"p"`
	OrderIds    []*DoDSettleOrder          `json:"orderIds,omitempty" msg:"o"`
}

type DoDSettleProduct struct {
	Seller    types.Address `json:"seller" msg:"s,extension"`
	ProductId string        `json:"productId,omitempty" msg:"p"`
}

func (z *DoDSettleProduct) Hash() types.Hash {
	data := append(z.Seller.Bytes(), []byte(z.ProductId)...)
	return types.HashData(data)
}

type DoDSettleOrder struct {
	Seller  types.Address `json:"seller" msg:"s,extension"`
	OrderId string        `json:"orderId,omitempty" msg:"o"`
}

func (z *DoDSettleOrder) Hash() types.Hash {
	data := append(z.Seller.Bytes(), []byte(z.OrderId)...)
	return types.HashData(data)
}

type DoDSettleCreateOrderParam struct {
	Buyer       *DoDSettleUser              `json:"buyer" msg:"b"`
	Seller      *DoDSettleUser              `json:"seller" msg:"s"`
	QuoteId     string                      `json:"quoteId,omitempty" msg:"q"`
	Connections []*DoDSettleConnectionParam `json:"connections,omitempty" msg:"c"`
}

func (z *DoDSettleCreateOrderParam) ToABI() ([]byte, error) {
	id := DoDSettlementABI.Methods[MethodNameDoDSettleCreateOrder].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *DoDSettleCreateOrderParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *DoDSettleCreateOrderParam) Verify() error {
	if z.Buyer == nil || z.Seller == nil || z.Connections == nil {
		return fmt.Errorf("invalid param")
	}

	quoteItemIdMap := make(map[string]struct{})
	for _, c := range z.Connections {
		if _, ok := quoteItemIdMap[c.QuoteItemId]; ok {
			return fmt.Errorf("duplicate quote item id")
		} else {
			quoteItemIdMap[c.QuoteItemId] = struct{}{}
		}
	}

	productItemIdMap := make(map[string]struct{})
	for _, c := range z.Connections {
		if _, ok := productItemIdMap[c.ItemId]; ok {
			return fmt.Errorf("duplicate product item id")
		} else {
			productItemIdMap[c.ItemId] = struct{}{}
		}
	}

	return nil
}

type DoDSettleConnectionParam struct {
	DoDSettleConnectionStaticParam
	DoDSettleConnectionDynamicParam
}

type DoDSettleConnectionStaticParam struct {
	ItemId         string `json:"itemId,omitempty" msg:"ii"`
	ProductId      string `json:"productId,omitempty" msg:"pi"`
	SrcCompanyName string `json:"srcCompanyName,omitempty" msg:"scn"`
	SrcRegion      string `json:"srcRegion,omitempty" msg:"sr"`
	SrcCity        string `json:"srcCity,omitempty" msg:"sc"`
	SrcDataCenter  string `json:"srcDataCenter,omitempty" msg:"sdc"`
	SrcPort        string `json:"srcPort,omitempty" msg:"sp"`
	DstCompanyName string `json:"dstCompanyName,omitempty" msg:"dcn"`
	DstRegion      string `json:"dstRegion,omitempty" msg:"dr"`
	DstCity        string `json:"dstCity,omitempty" msg:"dc"`
	DstDataCenter  string `json:"dstDataCenter,omitempty" msg:"ddc"`
	DstPort        string `json:"dstPort,omitempty" msg:"dp"`
}

type DoDSettleConnectionDynamicParam struct {
	OrderId        string                `json:"orderId,omitempty" msg:"oi"`
	QuoteItemId    string                `json:"quoteItemId,omitempty" msg:"qi"`
	ConnectionName string                `json:"connectionName,omitempty" msg:"cn"`
	PaymentType    DoDSettlePaymentType  `json:"paymentType,omitempty" msg:"pt"`
	BillingType    DoDSettleBillingType  `json:"billingType,omitempty" msg:"bt"`
	Currency       string                `json:"currency,omitempty" msg:"cr"`
	ServiceClass   DoDSettleServiceClass `json:"serviceClass,omitempty" msg:"scs"`
	Bandwidth      string                `json:"bandwidth,omitempty" msg:"bw"`
	BillingUnit    DoDSettleBillingUnit  `json:"billingUnit,omitempty" msg:"bu"`
	Price          float64               `json:"price,omitempty" msg:"p"`
	StartTime      int64                 `json:"startTime" msg:"st"`
	EndTime        int64                 `json:"endTime" msg:"et"`
}

type DoDSettleConnectionLifeTrack struct {
	OrderType DoDSettleOrderType               `json:"orderType,omitempty" msg:"ot"`
	OrderId   string                           `json:"orderId,omitempty" msg:"oi"`
	Time      int64                            `json:"time,omitempty" msg:"t"`
	Changed   *DoDSettleConnectionDynamicParam `json:"changed,omitempty" msg:"c"`
}

type DoDSettleConnectionInfo struct {
	DoDSettleConnectionStaticParam
	Active *DoDSettleConnectionDynamicParam   `json:"active" msg:"ac"`
	Done   []*DoDSettleConnectionDynamicParam `json:"done" msg:"do"`
	Track  []*DoDSettleConnectionLifeTrack    `json:"track" msg:"t"`
}

type DoDSettleOrderLifeTrack struct {
	ContractState DoDSettleContractState `json:"contractState" msg:"cs"`
	OrderState    DoDSettleOrderState    `json:"orderState" msg:"os"`
	Reason        string                 `json:"reason,omitempty" msg:"r"`
	Time          int64                  `json:"time" msg:"t"`
	Hash          types.Hash             `json:"hash" msg:"h,extension"`
}

type DoDSettleOrderInfo struct {
	Buyer         *DoDSettleUser              `json:"buyer" msg:"b"`
	Seller        *DoDSettleUser              `json:"seller" msg:"s"`
	OrderId       string                      `json:"orderId,omitempty" msg:"oi"`
	QuoteId       string                      `json:"quoteId,omitempty" msg:"qi"`
	OrderType     DoDSettleOrderType          `json:"orderType,omitempty" msg:"ot"`
	OrderState    DoDSettleOrderState         `json:"orderState" msg:"os"`
	ContractState DoDSettleContractState      `json:"contractState" msg:"cs"`
	Connections   []*DoDSettleConnectionParam `json:"connections" msg:"c"`
	Track         []*DoDSettleOrderLifeTrack  `json:"track" msg:"t"`
}

func NewOrderInfo() *DoDSettleOrderInfo {
	oi := new(DoDSettleOrderInfo)
	oi.Buyer = new(DoDSettleUser)
	oi.Seller = new(DoDSettleUser)
	oi.Connections = make([]*DoDSettleConnectionParam, 0)
	oi.Track = make([]*DoDSettleOrderLifeTrack, 0)
	return oi
}

type DoDSettleProductItem struct {
	ProductId string `json:"productId" msg:"p"`
	ItemId    string `json:"itemId" msg:"i"`
}

type DoDSettleUpdateOrderInfoParam struct {
	Buyer      types.Address           `json:"buyer" msg:"-"`
	InternalId types.Hash              `json:"internalId,omitempty" msg:"i,extension"`
	OrderId    string                  `json:"orderId,omitempty" msg:"oi"`
	ProductIds []*DoDSettleProductItem `json:"productIds" msg:"pis"`
	Status     DoDSettleOrderState     `json:"status,omitempty" msg:"s"`
	FailReason string                  `json:"failReason,omitempty" msg:"fr"`
}

func (z *DoDSettleUpdateOrderInfoParam) ToABI() ([]byte, error) {
	id := DoDSettlementABI.Methods[MethodNameDoDSettleUpdateOrderInfo].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *DoDSettleUpdateOrderInfoParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *DoDSettleUpdateOrderInfoParam) Verify(ctx *vmstore.VMContext) error {
	if z.ProductIds == nil {
		return fmt.Errorf("no product")
	}

	if z.InternalId.IsZero() {
		return fmt.Errorf("invalid internal id")
	}

	order, err := DoDSettleGetOrderInfoByInternalId(ctx, z.InternalId)
	if err != nil {
		return err
	}

	productIdMap := make(map[string]struct{})
	for _, p := range z.ProductIds {
		if _, ok := productIdMap[p.ItemId]; ok {
			return fmt.Errorf("duplicate product id")
		} else {
			productIdMap[p.ItemId] = struct{}{}
		}
	}

	if order.OrderType == DoDSettleOrderTypeCreate {
		for _, c := range order.Connections {
			found := false

			for _, p := range z.ProductIds {
				if c.ItemId == p.ItemId {
					found = true
					break
				}
			}

			if found == false {
				return fmt.Errorf("not enough products")
			}
		}
	}

	return nil
}

type DoDSettleChangeConnectionParam struct {
	ProductId   string `json:"productId" msg:"p"`
	QuoteItemId string `json:"quoteItemId" msg:"q"`
	DoDSettleConnectionDynamicParam
}

type DoDSettleChangeOrderParam struct {
	Buyer       *DoDSettleUser                    `json:"buyer" msg:"b"`
	Seller      *DoDSettleUser                    `json:"seller" msg:"s"`
	QuoteId     string                            `json:"quoteId" msg:"q"`
	Connections []*DoDSettleChangeConnectionParam `json:"connections" msg:"c"`
}

func (z *DoDSettleChangeOrderParam) ToABI() ([]byte, error) {
	id := DoDSettlementABI.Methods[MethodNameDoDSettleChangeOrder].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *DoDSettleChangeOrderParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *DoDSettleChangeOrderParam) Verify() error {
	if z.Buyer == nil || z.Seller == nil || z.Connections == nil {
		return fmt.Errorf("invalid param")
	}

	quoteItemIdMap := make(map[string]struct{})
	for _, c := range z.Connections {
		if _, ok := quoteItemIdMap[c.QuoteItemId]; ok {
			return fmt.Errorf("duplicate quote item id")
		} else {
			quoteItemIdMap[c.QuoteItemId] = struct{}{}
		}
	}

	return nil
}

type DoDSettleResponseParam struct {
	RequestHash types.Hash              `json:"requestHash" msg:"-"`
	Action      DoDSettleResponseAction `json:"action" msg:"c"`
}

type DoDSettleTerminateOrderParam struct {
	Buyer     *DoDSettleUser `json:"buyer" msg:"b"`
	Seller    *DoDSettleUser `json:"seller" msg:"s"`
	ProductId []string       `json:"productId" msg:"p"`
}

func (z *DoDSettleTerminateOrderParam) ToABI() ([]byte, error) {
	id := DoDSettlementABI.Methods[MethodNameDoDSettleTerminateOrder].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *DoDSettleTerminateOrderParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *DoDSettleTerminateOrderParam) Verify() error {
	if z.ProductId == nil {
		return fmt.Errorf("no product")
	}

	return nil
}

type DoDSettleResourceReadyParam struct {
	Address    types.Address `json:"address" msg:"-"`
	InternalId types.Hash    `json:"internalId" msg:"i,extension"`
	ProductId  []string      `json:"productId" msg:"p"`
}

func (z *DoDSettleResourceReadyParam) ToABI() ([]byte, error) {
	id := DoDSettlementABI.Methods[MethodNameDoDSettleResourceReady].Id()
	if data, err := z.MarshalMsg(nil); err != nil {
		return nil, err
	} else {
		id = append(id, data...)
		return id, nil
	}
}

func (z *DoDSettleResourceReadyParam) FromABI(data []byte) error {
	_, err := z.UnmarshalMsg(data[4:])
	return err
}

func (z *DoDSettleResourceReadyParam) Verify() error {
	return nil
}

type DoDSettleInvoiceConnDynamic struct {
	DoDSettleConnectionDynamicParam
	InvoiceStartTime int64   `json:"invoiceStartTime"`
	InvoiceEndTime   int64   `json:"invoiceEndTime"`
	InvoiceUnitCount int     `json:"invoiceUnitCount"`
	Amount           float64 `json:"amount"`
}

type DoDSettleInvoiceConnDetail struct {
	ConnectionAmount float64 `json:"connectionAmount"`
	DoDSettleConnectionStaticParam
	Usage []*DoDSettleInvoiceConnDynamic `json:"usage"`
}

type DoDSettleInvoiceOrderDetail struct {
	OrderId         string                        `json:"orderId"`
	ConnectionCount int                           `json:"connectionCount"`
	OrderAmount     float64                       `json:"orderAmount"`
	Connections     []*DoDSettleInvoiceConnDetail `json:"connections"`
}

type DoDSettleOrderInvoice struct {
	TotalConnectionCount int                          `json:"totalConnectionCount"`
	TotalAmount          float64                      `json:"totalAmount"`
	Currency             string                       `json:"currency"`
	StartTime            int64                        `json:"startTime"`
	EndTime              int64                        `json:"endTime"`
	Buyer                *DoDSettleUser               `json:"buyer"`
	Seller               *DoDSettleUser               `json:"seller"`
	Order                *DoDSettleInvoiceOrderDetail `json:"order"`
}

type DoDSettleBuyerInvoice struct {
	OrderCount           int                            `json:"orderCount"`
	TotalConnectionCount int                            `json:"totalConnectionCount"`
	TotalAmount          float64                        `json:"totalAmount"`
	Currency             string                         `json:"currency"`
	StartTime            int64                          `json:"startTime"`
	EndTime              int64                          `json:"endTime"`
	Buyer                *DoDSettleUser                 `json:"buyer"`
	Seller               *DoDSettleUser                 `json:"seller"`
	Orders               []*DoDSettleInvoiceOrderDetail `json:"orders"`
}

type DoDSettleProductInvoice struct {
	TotalAmount float64                     `json:"totalAmount"`
	Currency    string                      `json:"currency"`
	StartTime   int64                       `json:"startTime"`
	EndTime     int64                       `json:"endTime"`
	Buyer       *DoDSettleUser              `json:"buyer"`
	Seller      *DoDSettleUser              `json:"seller"`
	Connection  *DoDSettleInvoiceConnDetail `json:"connection"`
}

type DoDSettleConnectionActive struct {
	ActiveAt int64 `json:"activeAt" msg:"a"`
}
