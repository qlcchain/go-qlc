package abi

import (
	"fmt"

	"github.com/qlcchain/go-qlc/common/types"
)

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
CreateRequest = 1
CreateConfirmed
CreateSend
ChangeRequest
ChangeConfirmed
ChangeSend
TerminateRequest
TerminateConfirmed
TerminateSend
Rejected
Failed
Complete
)
*/
type DoDSettleOrderState int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
invoice = 1
stableCoin
)
*/
type DoDPaymentType int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
PAYG = 1
DOD
)
*/
type DoDBillingType int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
year = 1
month
week
day
hour
minute
second
)
*/
type DodBillingUnit int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
Gold = 1
Silver
Bronze
)
*/
type DoDServiceClass int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
confirm = 1
reject
)
*/
type DoDResponseAction int

//go:generate go-enum -f=$GOFILE --marshal --names
/*
ENUM(
create = 1
change
terminate
fail
)
*/
type DoDOrderOperation int

const (
	DoDSettleDBTableOrder uint8 = iota
	DoDSettleDBTableProduct
	DoDSettleDBTableIdMap
	DoDSettleDBTableUser
)

//go:generate msgp
type DoDSettleUser struct {
	Address types.Address `json:"address" msg:"a,extension"`
	Name    string        `json:"name" msg:"n"`
}

type DoDSettleUserInfo struct {
	InternalId types.Hash `json:"id" msg:"i,extension"`
}

type DoDSettleUserInfos struct {
	Infos []*DoDSettleUserInfo `json:"infos" msg:"i"`
}

type DoDSettleCreateOrderParam struct {
	Buyer       *DoDSettleUser              `json:"buyer" msg:"b"`
	Seller      *DoDSettleUser              `json:"seller" msg:"s"`
	Connections []*DoDSettleConnectionParam `json:"connections" msg:"c"`
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

	return nil
}

type DoDSettleConnectionParam struct {
	DoDSettleConnectionStaticParam
	DoDSettleConnectionDynamicParam
}

type DoDSettleConnectionStaticParam struct {
	ProductId      string `json:"productId" msg:"pi"`
	SrcCompanyName string `json:"srcCompanyName" msg:"scn"`
	SrcRegion      string `json:"srcRegion" msg:"sr"`
	SrcCity        string `json:"srcCity" msg:"sc"`
	SrcDataCenter  string `json:"srcDataCenter" msg:"sdc"`
	SrcPort        string `json:"srcPort" msg:"sp"`
	DstCompanyName string `json:"dstCompanyName" msg:"dcn"`
	DstRegion      string `json:"dstRegion" msg:"dr"`
	DstCity        string `json:"dstCity" msg:"dc"`
	DstDataCenter  string `json:"dstDataCenter" msg:"ddc"`
	DstPort        string `json:"dstPort" msg:"dp"`
}

type DoDSettleConnectionDynamicParam struct {
	ConnectionName string          `json:"connectionName" msg:"cn"`
	PaymentType    DoDPaymentType  `json:"paymentType" msg:"pt"`
	BillingType    DoDBillingType  `json:"billingType" msg:"bt"`
	Currency       string          `json:"currency" msg:"cr"`
	ServiceClass   DoDServiceClass `json:"serviceClass" msg:"scs"`
	Bandwidth      string          `json:"bandwidth" msg:"bw"`
	BillingUnit    DodBillingUnit  `json:"billingUnit" msg:"bu"`
	Price          float64         `json:"price" msg:"p"`
	StartTime      int64           `json:"startTime" msg:"st"`
	EndTime        int64           `json:"endTime" msg:"et"`
}

type DoDSettleConnectionLifeTrack struct {
	Param *DoDSettleConnectionDynamicParam
	Time  int64      `json:"time" msg:"t"`
	Hash  types.Hash `json:"hash" msg:"h,extension"`
}

type DoDSettleConnectionInfo struct {
	OrderId string `json:"orderId" msg:"oi"`
	Ready   bool   `json:"ready" msg:"r"`
	DoDSettleConnectionStaticParam
	Pending *DoDSettleConnectionDynamicParam   `json:"pending" msg:"pd"`
	Active  *DoDSettleConnectionDynamicParam   `json:"active" msg:"ac"`
	Done    []*DoDSettleConnectionDynamicParam `json:"done" msg:"do"`
	Track   []*DoDSettleConnectionLifeTrack    `json:"track" msg:"t"`
}

type DoDSettleOrderLifeTrack struct {
	State  DoDSettleOrderState `json:"state" msg:"o"`
	Reason string              `json:"reason" msg:"r"`
	Time   int64               `json:"time" msg:"t"`
	Hash   types.Hash          `json:"hash" msg:"h,extension"`
}

type DoDSettleOrderInfo struct {
	Buyer       *DoDSettleUser              `json:"buyer" msg:"b"`
	Seller      *DoDSettleUser              `json:"seller" msg:"s"`
	OrderId     string                      `json:"orderId" msg:"oi"`
	State       DoDSettleOrderState         `json:"state" msg:"st"`
	Connections []*DoDSettleConnectionParam `json:"connections" msg:"c"`
	Track       []*DoDSettleOrderLifeTrack  `json:"track" msg:"t"`
}

func NewOrderInfo() *DoDSettleOrderInfo {
	oi := new(DoDSettleOrderInfo)
	oi.Buyer = new(DoDSettleUser)
	oi.Seller = new(DoDSettleUser)
	oi.Connections = make([]*DoDSettleConnectionParam, 0)
	oi.Track = make([]*DoDSettleOrderLifeTrack, 0)
	return oi
}

type DoDSettleUpdateOrderInfoParam struct {
	Buyer      types.Address     `json:"buyer" msg:"-"`
	InternalId types.Hash        `json:"internalId" msg:"i,extension"`
	OrderId    string            `json:"orderId" msg:"oi"`
	ProductId  []string          `json:"productId" msg:"pi"`
	Operation  DoDOrderOperation `json:"operation" msg:"o"`
	FailReason string            `json:"failReason" msg:"fr"`
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

func (z *DoDSettleUpdateOrderInfoParam) Verify() error {
	return nil
}

type DoDSettleChangeConnectionParam struct {
	ProductId string `json:"productId" msg:"o"`
	DoDSettleConnectionDynamicParam
}

type DoDSettleChangeOrderParam struct {
	Buyer       *DoDSettleUser                    `json:"buyer" msg:"b"`
	Seller      *DoDSettleUser                    `json:"seller" msg:"s"`
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
	return nil
}

type DoDSettleResponseParam struct {
	RequestHash types.Hash        `json:"requestHash" msg:"-"`
	Action      DoDResponseAction `json:"action" msg:"c"`
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
	return nil
}

type DoDSettleResourceReadyParam struct {
	Address   types.Address `json:"address" msg:"-"`
	OrderId   string        `json:"orderId" msg:"o"`
	ProductId []string      `json:"productId" msg:"p"`
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
