package api

import (
	"fmt"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type DoDSettlementAPI struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	cc     *chainctx.ChainContext
	ctx    *vmstore.VMContext
	ca     *ContractApi
	co     *contract.DoDSettleCreateOrder
	uo     *contract.DoDSettleUpdateOrderInfo
	cho    *contract.DoDSettleChangeOrder
	to     *contract.DoDSettleTerminateOrder
	rr     *contract.DoDSettleResourceReady
}

func NewDoDSettlementAPI(cfgFile string, l ledger.Store) *DoDSettlementAPI {
	api := &DoDSettlementAPI{
		l:      l,
		logger: log.NewLogger("api dod settlement"),
		cc:     chainctx.NewChainContext(cfgFile),
		ctx:    vmstore.NewVMContext(l, &contractaddress.DoDSettlementAddress),
		co:     &contract.DoDSettleCreateOrder{},
		uo:     &contract.DoDSettleUpdateOrderInfo{},
		cho:    &contract.DoDSettleChangeOrder{},
		to:     &contract.DoDSettleTerminateOrder{},
		rr:     &contract.DoDSettleResourceReady{},
	}

	api.ca = NewContractApi(api.cc, api.l)
	return api
}

type DoDSettleCreateOrderParam struct {
	ContractPrivacyParam
	abi.DoDSettleCreateOrderParam
}

type DoDSettleResponseParam struct {
	ContractPrivacyParam
	abi.DoDSettleResponseParam
}

type DoDSettleUpdateOrderInfoParam struct {
	ContractPrivacyParam
	abi.DoDSettleUpdateOrderInfoParam
}

type DoDSettleChangeOrderParam struct {
	ContractPrivacyParam
	abi.DoDSettleChangeOrderParam
}

type DoDSettleTerminateOrderParam struct {
	ContractPrivacyParam
	abi.DoDSettleTerminateOrderParam
}

type DoDSettleResourceReadyParam struct {
	ContractPrivacyParam
	abi.DoDSettleResourceReadyParam
}

func (d *DoDSettlementAPI) GetCreateOrderBlock(param *DoDSettleCreateOrderParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	p := &ContractSendBlockPara{
		Address:        param.Buyer.Address,
		TokenName:      "QGAS",
		To:             contractaddress.DoDSettlementAddress,
		Amount:         types.NewBalance(0),
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateSendBlock(p)
}

func (d *DoDSettlementAPI) GetCreateOrderRewardBlock(param *DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	p := &ContractRewardBlockPara{
		SendHash:       param.RequestHash,
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateRewardBlock(p)
}

func (d *DoDSettlementAPI) GetUpdateOrderInfoBlock(param *DoDSettleUpdateOrderInfoParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	p := &ContractSendBlockPara{
		Address:        param.Buyer,
		TokenName:      "QGAS",
		To:             contractaddress.DoDSettlementAddress,
		Amount:         types.NewBalance(0),
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateSendBlock(p)
}

func (d *DoDSettlementAPI) GetUpdateOrderInfoRewardBlock(param *DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	sb, err := d.l.GetStateBlockConfirmed(param.RequestHash)
	if err != nil {
		return nil, fmt.Errorf("get request block err %s", err)
	}

	pm := new(abi.DoDSettleUpdateOrderInfoParam)
	err = pm.FromABI(sb.GetPayload())
	if err != nil {
		return nil, fmt.Errorf("unpack data err %s", err)
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, pm.InternalId)
	if err != nil {
		return nil, fmt.Errorf("get order by internalId err %s", pm.InternalId)
	}

	for _, p := range order.Connections {
		ak := &abi.DoDSettleConnectionActiveKey{InternalId: pm.InternalId, ProductId: p.ProductId}
		_, err := abi.DoDSettleGetSellerConnectionActive(d.ctx, ak.Hash())
		if err != nil {
			return nil, fmt.Errorf("product %s is not active, can't generate this reward", p.ProductId)
		}
	}

	p := &ContractRewardBlockPara{
		SendHash: param.RequestHash,
	}

	return d.ca.GenerateRewardBlock(p)
}

func (d *DoDSettlementAPI) GetChangeOrderBlock(param *DoDSettleChangeOrderParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	p := &ContractSendBlockPara{
		Address:        param.Buyer.Address,
		TokenName:      "QGAS",
		To:             contractaddress.DoDSettlementAddress,
		Amount:         types.NewBalance(0),
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateSendBlock(p)
}

func (d *DoDSettlementAPI) GetChangeOrderRewardBlock(param *DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	p := &ContractRewardBlockPara{
		SendHash:       param.RequestHash,
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateRewardBlock(p)
}

func (d *DoDSettlementAPI) GetTerminateOrderBlock(param *DoDSettleTerminateOrderParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	p := &ContractSendBlockPara{
		Address:        param.Buyer.Address,
		TokenName:      "QGAS",
		To:             contractaddress.DoDSettlementAddress,
		Amount:         types.NewBalance(0),
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateSendBlock(p)
}

func (d *DoDSettlementAPI) GetTerminateOrderRewardBlock(param *DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	p := &ContractRewardBlockPara{
		SendHash:       param.RequestHash,
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateRewardBlock(p)
}

func (d *DoDSettlementAPI) GetResourceReadyBlock(param *DoDSettleResourceReadyParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	p := &ContractSendBlockPara{
		Address:        param.Address,
		TokenName:      "QGAS",
		To:             contractaddress.DoDSettlementAddress,
		Amount:         types.NewBalance(0),
		Data:           data,
		PrivateFrom:    param.PrivateFrom,
		PrivateFor:     param.PrivateFor,
		PrivateGroupID: param.PrivateGroupID,
	}

	return d.ca.GenerateSendBlock(p)
}

func (d *DoDSettlementAPI) GetResourceReadyRewardBlock(param *DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	p := &ContractRewardBlockPara{
		SendHash: param.RequestHash,
	}

	return d.ca.GenerateRewardBlock(p)
}

func (d *DoDSettlementAPI) GetOrderInfoBySellerAndOrderId(seller types.Address, orderId string) (*abi.DoDSettleOrderInfo, error) {
	return abi.DoDSettleGetOrderInfoByOrderId(d.ctx, seller, orderId)
}

func (d *DoDSettlementAPI) GetOrderInfoByInternalId(internalId string) (*abi.DoDSettleOrderInfo, error) {
	id, err := types.NewHash(internalId)
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, id)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (d *DoDSettlementAPI) GetProductInfoBySellerAndProductId(seller types.Address, productId string) (*abi.DoDSettleConnectionInfo, error) {
	return abi.DoDSettleGetConnectionInfoByProductId(d.ctx, seller, productId)
}

type DoDPendingRequestRsp struct {
	Hash  types.Hash              `json:"hash"`
	Order *abi.DoDSettleOrderInfo `json:"order"`
}

// query all pending order requests
func (d *DoDSettlementAPI) GetPendingRequest(address types.Address) ([]*DoDPendingRequestRsp, error) {
	rsp := make([]*DoDPendingRequestRsp, 0)

	if err := d.l.GetPendingsByAddress(address, func(key *types.PendingKey, value *types.PendingInfo) error {
		sendBlock, err := d.ctx.GetStateBlockConfirmed(key.Hash)
		if err != nil {
			return err
		}

		if sendBlock.Type == types.ContractSend && sendBlock.Link == contractaddress.DoDSettlementAddress.ToHash() {
			method, err := abi.DoDSettlementABI.MethodById(sendBlock.Data)
			if err != nil {
				return err
			}

			if method.Name == abi.MethodNameDoDSettleCreateOrder || method.Name == abi.MethodNameDoDSettleChangeOrder ||
				method.Name == abi.MethodNameDoDSettleTerminateOrder {
				order, err := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, sendBlock.Previous)
				if err != nil {
					return err
				}

				r := &DoDPendingRequestRsp{
					Hash:  key.Hash,
					Order: order,
				}
				rsp = append(rsp, r)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return rsp, nil
}

type DoDSettleProductWithActiveInfo struct {
	ProductId string `json:"productId"`
	Active    bool   `json:"active"`
}

type DoDPendingResourceCheckInfo struct {
	SendHash   types.Hash                        `json:"sendHash"`
	OrderId    string                            `json:"orderId"`
	InternalId types.Hash                        `json:"internalId"`
	Products   []*DoDSettleProductWithActiveInfo `json:"products"`
}

// query all pending resource check requests
func (d *DoDSettlementAPI) GetPendingResourceCheck(address types.Address) ([]*DoDPendingResourceCheckInfo, error) {
	infos := make([]*DoDPendingResourceCheckInfo, 0)

	if err := d.l.GetPendingsByAddress(address, func(key *types.PendingKey, value *types.PendingInfo) error {
		sendBlock, err := d.ctx.GetStateBlockConfirmed(key.Hash)
		if err != nil {
			return err
		}

		if sendBlock.Type == types.ContractSend && sendBlock.Link == contractaddress.DoDSettlementAddress.ToHash() {
			method, err := abi.DoDSettlementABI.MethodById(sendBlock.Data)
			if err != nil {
				return err
			}

			if method.Name == abi.MethodNameDoDSettleUpdateOrderInfo {
				param := new(abi.DoDSettleUpdateOrderInfoParam)
				err = param.FromABI(sendBlock.Data)
				if err != nil {
					return err
				}

				if param.Status == abi.DoDSettleOrderStateFail {
					return nil
				}

				info := &DoDPendingResourceCheckInfo{
					SendHash:   key.Hash,
					OrderId:    param.OrderId,
					InternalId: param.InternalId,
					Products:   make([]*DoDSettleProductWithActiveInfo, 0),
				}

				order, err := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, param.InternalId)
				if err != nil {
					return nil
				}

				for _, p := range order.Connections {
					pai := &DoDSettleProductWithActiveInfo{ProductId: p.ProductId, Active: false}

					ak := &abi.DoDSettleConnectionActiveKey{InternalId: param.InternalId, ProductId: p.ProductId}
					_, err := abi.DoDSettleGetSellerConnectionActive(d.ctx, ak.Hash())
					if err == nil {
						pai.Active = true
					}

					info.Products = append(info.Products, pai)
				}

				infos = append(infos, info)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return infos, nil
}

type DoDPlacingOrderInfo struct {
	InternalId types.Hash              `json:"internalId"`
	OrderInfo  *abi.DoDSettleOrderInfo `json:"orderInfo"`
}

type DoDPlacingOrderResp struct {
	TotalOrders int                    `json:"totalOrders"`
	OrderList   []*DoDPlacingOrderInfo `json:"orderList"`
}

func (d *DoDSettlementAPI) GetPlacingOrder(buyer, seller types.Address, count, offset int) (*DoDPlacingOrderResp, error) {
	resp := new(DoDPlacingOrderResp)
	resp.OrderList = make([]*DoDPlacingOrderInfo, 0)
	ol := make([]*DoDPlacingOrderInfo, 0)

	internalIds, err := abi.DoDSettleGetInternalIdListByAddress(d.ctx, buyer)
	if err != nil {
		return nil, err
	}

	for _, id := range internalIds {
		oi, _ := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, id)
		if oi == nil {
			continue
		}

		if oi.Seller.Address != seller {
			continue
		}

		if oi.ContractState == abi.DoDSettleContractStateRequest ||
			(oi.ContractState == abi.DoDSettleContractStateConfirmed && oi.OrderState == abi.DoDSettleOrderStateNull) {
			poi := &DoDPlacingOrderInfo{
				InternalId: id,
				OrderInfo:  oi,
			}
			ol = append(ol, poi)
		}
	}

	resp.TotalOrders = len(ol)

	for i, o := range ol {
		if i+1 <= offset {
			continue
		}

		if i+1 > offset+count {
			break
		}

		resp.OrderList = append(resp.OrderList, o)
	}

	return resp, nil
}

func (d *DoDSettlementAPI) GetProductIdListByAddress(address types.Address) ([]*abi.DoDSettleProduct, error) {
	return abi.DoDSettleGetProductIdListByAddress(d.ctx, address)
}

func (d *DoDSettlementAPI) GetOrderIdListByAddress(address types.Address) ([]*abi.DoDSettleOrder, error) {
	return abi.DoDSettleGetOrderIdListByAddress(d.ctx, address)
}

func (d *DoDSettlementAPI) GetProductIdListByAddressAndSeller(address, seller types.Address) ([]*abi.DoDSettleProduct, error) {
	all, err := abi.DoDSettleGetProductIdListByAddress(d.ctx, address)
	if err != nil {
		return nil, err
	}

	products := make([]*abi.DoDSettleProduct, 0)
	for _, p := range all {
		if p.Seller == seller {
			products = append(products, p)
		}
	}

	return products, nil
}

func (d *DoDSettlementAPI) GetOrderIdListByAddressAndSeller(address, seller types.Address) ([]*abi.DoDSettleOrder, error) {
	all, err := abi.DoDSettleGetOrderIdListByAddress(d.ctx, address)
	if err != nil {
		return nil, err
	}

	orders := make([]*abi.DoDSettleOrder, 0)
	for _, p := range all {
		if p.Seller == seller {
			orders = append(orders, p)
		}
	}

	return orders, nil
}

type DoDSettlementOrderInfoResp struct {
	OrderInfo   []*abi.DoDSettleOrderInfo `json:"orderInfo"`
	TotalOrders int                       `json:"totalOrders"`
}

func (d *DoDSettlementAPI) GetOrderCountByAddress(address types.Address) int {
	order, err := abi.DoDSettleGetInternalIdListByAddress(d.ctx, address)
	if err != nil {
		return 0
	} else {
		return len(order)
	}
}

func (d *DoDSettlementAPI) GetOrderInfoByAddress(address types.Address, count, offset int) (*DoDSettlementOrderInfoResp, error) {
	orders := make([]*abi.DoDSettleOrderInfo, 0)

	internalIds, err := abi.DoDSettleGetInternalIdListByAddress(d.ctx, address)
	if err != nil {
		return nil, err
	}

	index := 0

	for i := len(internalIds) - 1; i >= 0; i-- {
		index++

		if index <= offset {
			continue
		}

		if index > offset+count {
			break
		}

		order, err := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, internalIds[i])
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return &DoDSettlementOrderInfoResp{OrderInfo: orders, TotalOrders: len(internalIds)}, nil
}

func (d *DoDSettlementAPI) GetOrderCountByAddressAndSeller(address, seller types.Address) int {
	order, err := d.GetOrderIdListByAddressAndSeller(address, seller)
	if err != nil {
		return 0
	} else {
		return len(order)
	}
}

func (d *DoDSettlementAPI) GetOrderInfoByAddressAndSeller(address, seller types.Address, count, offset int) (*DoDSettlementOrderInfoResp, error) {
	orders := make([]*abi.DoDSettleOrderInfo, 0)

	orderIds, err := d.GetOrderIdListByAddressAndSeller(address, seller)
	if err != nil {
		return nil, err
	}

	index := 0

	for i := len(orderIds) - 1; i >= 0; i-- {
		index++

		if index <= offset {
			continue
		}

		if index > offset+count {
			break
		}

		order, err := abi.DoDSettleGetOrderInfoByOrderId(d.ctx, seller, orderIds[i].OrderId)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return &DoDSettlementOrderInfoResp{OrderInfo: orders, TotalOrders: len(orderIds)}, nil
}

type DoDSettlementProductInfoResp struct {
	ProductInfo   []*abi.DoDSettleConnectionInfo `json:"productInfo"`
	TotalProducts int                            `json:"totalProducts"`
}

func (d *DoDSettlementAPI) GetProductCountByAddress(address types.Address) int {
	product, err := abi.DoDSettleGetProductIdListByAddress(d.ctx, address)
	if err != nil {
		return 0
	} else {
		return len(product)
	}
}

func (d *DoDSettlementAPI) GetProductInfoByAddress(address types.Address, count, offset int) (*DoDSettlementProductInfoResp, error) {
	products := make([]*abi.DoDSettleConnectionInfo, 0)

	productIds, err := abi.DoDSettleGetProductIdListByAddress(d.ctx, address)
	if err != nil {
		return nil, err
	}

	index := 0

	for i := len(productIds) - 1; i >= 0; i-- {
		index++

		if index <= offset {
			continue
		}

		if index > offset+count {
			break
		}

		product, err := abi.DoDSettleGetConnectionInfoByProductId(d.ctx, productIds[i].Seller, productIds[i].ProductId)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return &DoDSettlementProductInfoResp{ProductInfo: products, TotalProducts: len(productIds)}, nil
}

func (d *DoDSettlementAPI) GetProductCountByAddressAndSeller(address, seller types.Address) int {
	product, err := d.GetProductIdListByAddressAndSeller(address, seller)
	if err != nil {
		return 0
	} else {
		return len(product)
	}
}

func (d *DoDSettlementAPI) GetProductInfoByAddressAndSeller(address, seller types.Address, count, offset int) (*DoDSettlementProductInfoResp, error) {
	products := make([]*abi.DoDSettleConnectionInfo, 0)

	productIds, err := d.GetProductIdListByAddressAndSeller(address, seller)
	if err != nil {
		return nil, err
	}

	index := 0

	for i := len(productIds) - 1; i >= 0; i-- {
		index++

		if index <= offset {
			continue
		}

		if index > offset+count {
			break
		}

		product, err := abi.DoDSettleGetConnectionInfoByProductId(d.ctx, seller, productIds[i].ProductId)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	return &DoDSettlementProductInfoResp{ProductInfo: products, TotalProducts: len(productIds)}, nil
}

func (d *DoDSettlementAPI) GenerateInvoiceByOrderId(seller types.Address, orderId string, start, end int64, flight, split bool) (*abi.DoDSettleOrderInvoice, error) {
	return abi.DoDSettleGenerateInvoiceByOrder(d.ctx, seller, orderId, start, end, flight, split)
}

func (d *DoDSettlementAPI) GenerateInvoiceByBuyer(seller, buyer types.Address, start, end int64, flight, split bool) (*abi.DoDSettleBuyerInvoice, error) {
	return abi.DoDSettleGenerateInvoiceByBuyer(d.ctx, seller, buyer, start, end, flight, split)
}

func (d *DoDSettlementAPI) GenerateInvoiceByProductId(seller types.Address, productId string, start, end int64, flight, split bool) (*abi.DoDSettleProductInvoice, error) {
	return abi.DoDSettleGenerateInvoiceByProduct(d.ctx, seller, productId, start, end, flight, split)
}
