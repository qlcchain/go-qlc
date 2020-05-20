package api

import (
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

	return abi.DoDSettleGetOrderInfoByInternalId(d.ctx, id)
}

func (d *DoDSettlementAPI) GetConnectionInfoBySellerAndProductId(seller types.Address, productId string) (*abi.DoDSettleConnectionInfo, error) {
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
	OrderId  string                            `json:"orderId"`
	Products []*DoDSettleProductWithActiveInfo `json:"products"`
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

				order, err := abi.DoDSettleGetOrderInfoByInternalId(d.ctx, param.InternalId)
				if err != nil {
					return err
				}

				info := &DoDPendingResourceCheckInfo{
					OrderId:  param.OrderId,
					Products: make([]*DoDSettleProductWithActiveInfo, 0),
				}

				for _, pid := range param.ProductId {
					productKey := &abi.DoDSettleProduct{
						Seller:    order.Seller.Address,
						ProductId: pid,
					}

					pai := &DoDSettleProductWithActiveInfo{ProductId: pid, Active: false}

					_, err := abi.DodSettleGetSellerConnectionActive(d.ctx, productKey.Hash())
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

func (d *DoDSettlementAPI) GetPlacingOrder(buyer, seller types.Address) ([]*DoDPlacingOrderInfo, error) {
	orderInfo := make([]*DoDPlacingOrderInfo, 0)

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

		if oi.ContractState == abi.DoDSettleContractStateConfirmed && oi.OrderState == abi.DoDSettleOrderStateNull {
			poi := &DoDPlacingOrderInfo{
				InternalId: id,
				OrderInfo:  oi,
			}
			orderInfo = append(orderInfo, poi)
		}
	}

	return orderInfo, nil
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

func (d *DoDSettlementAPI) GenerateInvoiceByOrderId(seller types.Address, orderId string, start, end int64) (*abi.DoDSettleInvoice, error) {
	return abi.DodSettleGenerateInvoiceByOrder(d.ctx, seller, orderId, start, end)
}

func (d *DoDSettlementAPI) GenerateInvoiceByBuyer(seller, buyer types.Address, start, end int64) (*abi.DoDSettleInvoice, error) {
	return abi.DodSettleGenerateInvoiceByBuyer(d.ctx, seller, buyer, start, end)
}
