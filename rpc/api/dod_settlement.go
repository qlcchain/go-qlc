package api

import (
	"fmt"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/config"
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
	return api
}

func (d *DoDSettlementAPI) GetCreateOrderBlock(param *abi.DoDSettleCreateOrderParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Buyer.Address)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Buyer.Address)
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Buyer.Address,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDSettlementAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	_, _, err = d.co.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (d *DoDSettlementAPI) GetCreateOrderRewardBlock(param *abi.DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	send, err := d.l.GetStateBlockConfirmed(param.RequestHash)
	if err != nil {
		return nil, err
	}

	reward := &types.StateBlock{}
	reward.Data, err = param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	blocks, err := d.co.DoReceive(vmContext, reward, send)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		h := vmstore.TrieHash(vmContext)
		if h != nil {
			reward.Extra = *h
		}

		return reward, nil
	}

	return reward, nil
}

func (d *DoDSettlementAPI) GetUpdateOrderInfoBlock(param *abi.DoDSettleUpdateOrderInfoParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Buyer)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Buyer)
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Buyer,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDSettlementAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	_, _, err = d.uo.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (d *DoDSettlementAPI) GetUpdateOrderInfoRewardBlock(param *abi.DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	send, err := d.l.GetStateBlockConfirmed(param.RequestHash)
	if err != nil {
		return nil, err
	}

	reward := &types.StateBlock{}
	reward.Data, err = param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	blocks, err := d.uo.DoReceive(vmContext, reward, send)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		h := vmstore.TrieHash(vmContext)
		if h != nil {
			reward.Extra = *h
		}

		return reward, nil
	}

	return reward, nil
}

func (d *DoDSettlementAPI) GetChangeOrderBlock(param *abi.DoDSettleChangeOrderParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Buyer.Address)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Buyer.Address)
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Buyer.Address,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDSettlementAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	_, _, err = d.cho.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (d *DoDSettlementAPI) GetChangeOrderRewardBlock(param *abi.DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	send, err := d.l.GetStateBlockConfirmed(param.RequestHash)
	if err != nil {
		return nil, err
	}

	reward := &types.StateBlock{}
	reward.Data, err = param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	blocks, err := d.cho.DoReceive(vmContext, reward, send)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		h := vmstore.TrieHash(vmContext)
		if h != nil {
			reward.Extra = *h
		}

		return reward, nil
	}

	return reward, nil
}

func (d *DoDSettlementAPI) GetTerminateOrderBlock(param *abi.DoDSettleTerminateOrderParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Buyer.Address)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Buyer.Address)
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Buyer.Address,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDSettlementAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	_, _, err = d.to.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (d *DoDSettlementAPI) GetTerminateOrderRewardBlock(param *abi.DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	send, err := d.l.GetStateBlockConfirmed(param.RequestHash)
	if err != nil {
		return nil, err
	}

	reward := &types.StateBlock{}
	reward.Data, err = param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	blocks, err := d.to.DoReceive(vmContext, reward, send)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		h := vmstore.TrieHash(vmContext)
		if h != nil {
			reward.Extra = *h
		}

		return reward, nil
	}

	return reward, nil
}

func (d *DoDSettlementAPI) GetResourceReadyBlock(param *abi.DoDSettleResourceReadyParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !d.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	povHeader, err := d.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := d.l.GetAccountMeta(param.Address)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.GasToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Address)
	}

	data, err := param.ToABI()
	if err != nil {
		return nil, err
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Address,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(contractaddress.DoDSettlementAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	_, _, err = d.rr.ProcessSend(vmContext, send)
	if err != nil {
		return nil, err
	}

	h := vmstore.TrieHash(vmContext)
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (d *DoDSettlementAPI) GetResourceReadyRewardBlock(param *abi.DoDSettleResponseParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	send, err := d.l.GetStateBlockConfirmed(param.RequestHash)
	if err != nil {
		return nil, err
	}

	reward := &types.StateBlock{}
	reward.Data, err = param.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	vmContext := vmstore.NewVMContext(d.l, &contractaddress.DoDSettlementAddress)
	blocks, err := d.rr.DoReceive(vmContext, reward, send)
	if err != nil {
		return nil, err
	}

	if len(blocks) > 0 {
		h := vmstore.TrieHash(vmContext)
		if h != nil {
			reward.Extra = *h
		}

		return reward, nil
	}

	return reward, nil
}

func (d *DoDSettlementAPI) GetOrderInfoByOrderId(orderId string) (*abi.DoDSettleOrderInfo, error) {
	return abi.DoDSettleGetOrderInfoByOrderId(d.ctx, orderId)
}

func (d *DoDSettlementAPI) GetOrderInfoByInternalId(internalId string) (*abi.DoDSettleOrderInfo, error) {
	id, err := types.NewHash(internalId)
	if err != nil {
		return nil, err
	}

	return abi.DoDSettleGetOrderInfoByInternalId(d.ctx, id)
}

func (d *DoDSettlementAPI) GetConnectionInfoByProductId(productId string) (*abi.DoDSettleConnectionInfo, error) {
	productIdHash := types.Sha256DHashData([]byte(productId))
	return abi.DoDSettleGetConnectionInfoByProductId(d.ctx, productIdHash)
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

// query all pending resource check requests
func (d *DoDSettlementAPI) GetPendingResourceCheck(address types.Address) ([]string, error) {
	orderId := make([]string, 0)

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

				orderId = append(orderId, param.OrderId)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return orderId, nil
}
