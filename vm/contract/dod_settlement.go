package contract

import (
	"fmt"
	"time"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var DoDSettlementContract = NewChainContract(
	map[string]Contract{
		abi.MethodNameDoDSettleCreateOrder: &DoDSettleCreateOrder{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					pending:   true,
				},
			},
		},
		abi.MethodNameDoDSettleUpdateOrderInfo: &DoDSettleUpdateOrderInfo{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					pending:   true,
				},
			},
		},
		abi.MethodNameDoDSettleChangeOrder: &DoDSettleChangeOrder{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					pending:   true,
				},
			},
		},
		abi.MethodNameDoDSettleTerminateOrder: &DoDSettleTerminateOrder{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					pending:   true,
				},
			},
		},
		abi.MethodNameDoDSettleResourceReady: &DoDSettleResourceReady{
			BaseContract: BaseContract{
				Describe: Describe{
					specVer:   SpecVer2,
					signature: true,
					work:      true,
					pending:   true,
				},
			},
		},
	},
	abi.DoDSettlementABI,
	abi.JsonDoDSettlement,
)

type DoDSettleCreateOrder struct {
	BaseContract
}

func (co *DoDSettleCreateOrder) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	_, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken())
	if err != nil {
		return nil, nil, err
	}

	param := new(abi.DoDSettleCreateOrderParam)
	err = param.FromABI(block.GetPayload())
	if err != nil {
		return nil, nil, err
	}

	err = param.Verify()
	if err != nil {
		return nil, nil, err
	}

	if param.Buyer.Address != block.Address {
		return nil, nil, fmt.Errorf("invalid operator")
	}

	err = co.setStorage(ctx, param, block)
	if err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Seller.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: param.Buyer.Address,
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func (co *DoDSettleCreateOrder) setStorage(ctx *vmstore.VMContext, param *abi.DoDSettleCreateOrderParam, block *types.StateBlock) error {
	order := abi.NewOrderInfo()

	order.Buyer.Address = param.Buyer.Address
	order.Buyer.Name = param.Buyer.Name
	order.Seller.Address = param.Seller.Address
	order.Seller.Name = param.Seller.Name
	order.OrderType = abi.DoDSettleOrderTypeCreate
	order.ContractState = abi.DoDSettleContractStateRequest

	for _, c := range param.Connections {
		conn := &abi.DoDSettleConnectionParam{
			DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
				SrcCompanyName: c.SrcCompanyName,
				SrcRegion:      c.SrcRegion,
				SrcCity:        c.SrcCity,
				SrcDataCenter:  c.SrcDataCenter,
				SrcPort:        c.SrcPort,
				DstCompanyName: c.DstCompanyName,
				DstRegion:      c.DstRegion,
				DstCity:        c.DstCity,
				DstDataCenter:  c.DstDataCenter,
				DstPort:        c.DstPort,
			},
			DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
				ConnectionName: c.ConnectionName,
				PaymentType:    c.PaymentType,
				BillingType:    c.BillingType,
				Currency:       c.Currency,
				ServiceClass:   c.ServiceClass,
				Bandwidth:      c.Bandwidth,
				BillingUnit:    c.BillingUnit,
				Price:          c.Price,
				StartTime:      c.StartTime,
				EndTime:        c.EndTime,
			},
		}
		order.Connections = append(order.Connections, conn)
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	data, err := order.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableOrder)
	key = append(key, block.Previous.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	key = key[0:0]
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, param.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)

	data, err = ctx.GetStorage(nil, key)
	if err != nil {
		userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
		userInfo.ProductIds = make([]*abi.DoDSettleProduct, 0)
		userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

		internalId := &abi.DoDSettleInternalIdWrap{InternalId: block.Previous}
		userInfo.InternalIds = append(userInfo.InternalIds, internalId)
	} else {
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		internalId := &abi.DoDSettleInternalIdWrap{InternalId: block.Previous}
		userInfo.InternalIds = append(userInfo.InternalIds, internalId)
	}

	data, err = userInfo.MarshalMsg(nil)
	if err != nil {
		return err
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}

func (co *DoDSettleCreateOrder) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(abi.DoDSettleResponseParam)
	_, err := param.UnmarshalMsg(block.GetPayload())
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, input.Previous)
	if err != nil {
		return nil, err
	}

	if param.Action == abi.DoDSettleResponseActionConfirm {
		order.ContractState = abi.DoDSettleContractStateConfirmed
	} else {
		order.ContractState = abi.DoDSettleContractStateRejected
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = order.Seller.Address
	block.Token = cfg.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.NewBalance(0)
	block.Oracle = types.NewBalance(0)
	block.Storage = types.NewBalance(0)
	block.Network = types.NewBalance(0)

	am, _ := ctx.GetAccountMeta(block.Address)
	if am != nil {
		tm := am.Token(cfg.GasToken())
		if tm != nil {
			block.Balance = tm.Balance
			block.Representative = tm.Representative
			block.Previous = tm.Header
		} else {
			block.Balance = types.NewBalance(0)
			if len(am.Tokens) > 0 {
				block.Representative = am.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = types.NewBalance(0)
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	err = abi.DoDSettleUpdateOrder(ctx, order, input.Previous)
	if err != nil {
		return nil, err
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: order.Seller.Address,
			BlockType: types.ContractReward,
			Amount:    types.NewBalance(0),
			Token:     cfg.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (co *DoDSettleCreateOrder) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	param := new(abi.DoDSettleCreateOrderParam)
	err := param.FromABI(block.GetPayload())
	if err != nil {
		return types.ZeroAddress, err
	}

	return param.Seller.Address, nil
}

type DoDSettleUpdateOrderInfo struct {
	BaseContract
}

func (uo *DoDSettleUpdateOrderInfo) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	_, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken())
	if err != nil {
		return nil, nil, err
	}

	param := new(abi.DoDSettleUpdateOrderInfoParam)
	err = param.FromABI(block.GetPayload())
	if err != nil {
		return nil, nil, err
	}

	err = param.Verify()
	if err != nil {
		return nil, nil, err
	}

	if param.InternalId.IsZero() {
		return nil, nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, param.InternalId)
	if err != nil {
		return nil, nil, err
	}

	if order.Buyer.Address != block.Address {
		return nil, nil, fmt.Errorf("invalid operator")
	}

	order.OrderId = param.OrderId
	order.OrderState = param.Status

	// maybe not matched ? need to check this.
	if order.OrderType == abi.DoDSettleOrderTypeCreate {
		for i, c := range order.Connections {
			c.ProductId = param.ProductId[i]
		}
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		OrderState:    order.OrderState,
		Reason:        param.FailReason,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	err = uo.setStorage(ctx, order, param.InternalId, true)
	if err != nil {
		return nil, nil, err
	}

	if order.OrderState != abi.DoDSettleOrderStateFail {
		for _, cp := range order.Connections {
			var conn *abi.DoDSettleConnectionInfo

			productKey := &abi.DoDSettleProduct{
				Seller:    order.Seller.Address,
				ProductId: cp.ProductId,
			}
			productHash := productKey.Hash()

			// only update dod
			if cp.BillingType == abi.DoDSettleBillingTypePAYG {
				continue
			}

			if order.OrderType == abi.DoDSettleOrderTypeCreate {
				conn, _ = abi.DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
				if conn != nil {
					return nil, nil, fmt.Errorf("dup product")
				}

				conn = &abi.DoDSettleConnectionInfo{
					DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
						ProductId:      cp.ProductId,
						SrcCompanyName: cp.SrcCompanyName,
						SrcRegion:      cp.SrcRegion,
						SrcCity:        cp.SrcCity,
						SrcDataCenter:  cp.SrcDataCenter,
						SrcPort:        cp.SrcPort,
						DstCompanyName: cp.DstCompanyName,
						DstRegion:      cp.DstRegion,
						DstCity:        cp.DstCity,
						DstDataCenter:  cp.DstDataCenter,
						DstPort:        cp.DstPort,
					},
					Active: &abi.DoDSettleConnectionDynamicParam{
						ConnectionName: cp.ConnectionName,
						PaymentType:    cp.PaymentType,
						BillingType:    cp.BillingType,
						Currency:       cp.Currency,
						ServiceClass:   cp.ServiceClass,
						Bandwidth:      cp.Bandwidth,
						BillingUnit:    cp.BillingUnit,
						Price:          cp.Price,
						StartTime:      cp.StartTime,
						EndTime:        cp.EndTime,
					},
					Done:  make([]*abi.DoDSettleConnectionDynamicParam, 0),
					Track: make([]*abi.DoDSettleConnectionLifeTrack, 0),
				}
			} else if order.OrderType == abi.DoDSettleOrderTypeChange {
				conn, err = abi.DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
				if err != nil {
					return nil, nil, err
				}

				newActive := &abi.DoDSettleConnectionDynamicParam{
					ConnectionName: cp.ConnectionName,
					PaymentType:    cp.PaymentType,
					BillingType:    cp.BillingType,
					Currency:       cp.Currency,
					ServiceClass:   cp.ServiceClass,
					Bandwidth:      cp.Bandwidth,
					BillingUnit:    cp.BillingUnit,
					Price:          cp.Price,
					StartTime:      cp.StartTime,
					EndTime:        cp.EndTime,
				}

				abi.DoDSettleInheritParam(conn.Active, newActive)

				conn.Done = append(conn.Done, conn.Active)
				conn.Active = newActive
			} else {
				conn, err = abi.DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
				if err != nil {
					return nil, nil, err
				}

				conn.Done = append(conn.Done, conn.Active)
				conn.Active = nil
			}

			track := &abi.DoDSettleConnectionLifeTrack{
				OrderType: order.OrderType,
				OrderId:   order.OrderId,
				Time:      time.Now().Unix(),
			}

			if order.OrderType == abi.DoDSettleOrderTypeCreate || order.OrderType == abi.DoDSettleOrderTypeChange {
				track.Changed = &abi.DoDSettleConnectionDynamicParam{
					ConnectionName: cp.ConnectionName,
					PaymentType:    cp.PaymentType,
					BillingType:    cp.BillingType,
					Currency:       cp.Currency,
					ServiceClass:   cp.ServiceClass,
					Bandwidth:      cp.Bandwidth,
					BillingUnit:    cp.BillingUnit,
					Price:          cp.Price,
					StartTime:      cp.StartTime,
					EndTime:        cp.EndTime,
				}
			}

			conn.Track = append(conn.Track, track)

			err = abi.DoDSettleUpdateConnection(ctx, conn, productHash)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return &types.PendingKey{
			Address: order.Seller.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: order.Buyer.Address,
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func (uo *DoDSettleUpdateOrderInfo) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	param := new(abi.DoDSettleUpdateOrderInfoParam)
	err := param.FromABI(block.GetPayload())
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	if param.InternalId.IsZero() {
		return common.ContractNoGap, nil, fmt.Errorf("invalid internal id")
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, param.InternalId)
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	if order.ContractState != abi.DoDSettleContractStateConfirmed {
		return common.ContractDoDOrderState, param.InternalId, nil
	}

	return common.ContractNoGap, nil, nil
}

func (uo *DoDSettleUpdateOrderInfo) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(abi.DoDSettleUpdateOrderInfoParam)
	err := param.FromABI(input.GetPayload())
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, param.InternalId)
	if err != nil {
		return nil, err
	}

	if order.OrderState == abi.DoDSettleOrderStateSuccess {
		order.OrderState = abi.DoDSettleOrderStateComplete
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = order.Seller.Address
	block.Token = cfg.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.NewBalance(0)
	block.Oracle = types.NewBalance(0)
	block.Storage = types.NewBalance(0)
	block.Network = types.NewBalance(0)

	am, _ := ctx.GetAccountMeta(block.Address)
	if am != nil {
		tm := am.Token(cfg.GasToken())
		if tm != nil {
			block.Balance = tm.Balance
			block.Representative = tm.Representative
			block.Previous = tm.Header
		} else {
			block.Balance = types.NewBalance(0)
			if len(am.Tokens) > 0 {
				block.Representative = am.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = types.NewBalance(0)
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		OrderState:    order.OrderState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	err = uo.setStorage(ctx, order, param.InternalId, false)
	if err != nil {
		return nil, err
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: order.Seller.Address,
			BlockType: types.ContractReward,
			Amount:    types.NewBalance(0),
			Token:     cfg.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (uo *DoDSettleUpdateOrderInfo) setStorage(ctx *vmstore.VMContext, order *abi.DoDSettleOrderInfo, id types.Hash, updateOrderId bool) error {
	data, err := order.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableOrder)
	key = append(key, id.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	if updateOrderId {
		orderKey := &abi.DoDSettleOrder{
			Seller:  order.Seller.Address,
			OrderId: order.OrderId,
		}

		key = key[0:0]
		key = append(key, abi.DoDSettleDBTableOrderIdMap)
		key = append(key, orderKey.Hash().Bytes()...)

		err = ctx.SetStorage(nil, key, id.Bytes())
		if err != nil {
			return err
		}

		key = key[0:0]
		key = append(key, abi.DoDSettleDBTableUser)
		key = append(key, order.Buyer.Address.Bytes()...)

		data, err = ctx.GetStorage(nil, key)
		if err != nil {
			return err
		}

		userInfo := new(abi.DoDSettleUserInfos)
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		userInfo.OrderIds = append(userInfo.OrderIds, orderKey)

		if order.OrderType == abi.DoDSettleOrderTypeCreate {
			for _, c := range order.Connections {
				productKey := &abi.DoDSettleProduct{
					Seller:    order.Seller.Address,
					ProductId: c.ProductId,
				}

				userInfo.ProductIds = append(userInfo.ProductIds, productKey)
			}
		}

		data, err = userInfo.MarshalMsg(nil)
		if err != nil {
			return err
		}

		err = ctx.SetStorage(nil, key, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uo *DoDSettleUpdateOrderInfo) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	param := new(abi.DoDSettleUpdateOrderInfoParam)
	err := param.FromABI(block.GetPayload())
	if err != nil {
		return types.ZeroAddress, err
	}

	if param.InternalId.IsZero() {
		return types.ZeroAddress, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, param.InternalId)
	if err != nil {
		return types.ZeroAddress, err
	}

	return order.Seller.Address, nil
}

type DoDSettleChangeOrder struct {
	BaseContract
}

func (co *DoDSettleChangeOrder) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	_, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken())
	if err != nil {
		return nil, nil, err
	}

	param := new(abi.DoDSettleChangeOrderParam)
	err = param.FromABI(block.GetPayload())
	if err != nil {
		return nil, nil, err
	}

	err = param.Verify()
	if err != nil {
		return nil, nil, err
	}

	if param.Buyer.Address != block.Address {
		return nil, nil, fmt.Errorf("invalid operator")
	}

	err = co.setStorage(ctx, param, block)
	if err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Seller.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: param.Buyer.Address,
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func (co *DoDSettleChangeOrder) setStorage(ctx *vmstore.VMContext, param *abi.DoDSettleChangeOrderParam, block *types.StateBlock) error {
	order := abi.NewOrderInfo()

	order.Buyer.Address = param.Buyer.Address
	order.Buyer.Name = param.Buyer.Name
	order.Seller.Address = param.Seller.Address
	order.Seller.Name = param.Seller.Name
	order.OrderType = abi.DoDSettleOrderTypeChange
	order.ContractState = abi.DoDSettleContractStateRequest

	for _, c := range param.Connections {
		conn := &abi.DoDSettleConnectionParam{
			DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
				ProductId: c.ProductId,
			},
			DoDSettleConnectionDynamicParam: abi.DoDSettleConnectionDynamicParam{
				ConnectionName: c.ConnectionName,
				PaymentType:    c.PaymentType,
				BillingType:    c.BillingType,
				Currency:       c.Currency,
				ServiceClass:   c.ServiceClass,
				Bandwidth:      c.Bandwidth,
				BillingUnit:    c.BillingUnit,
				Price:          c.Price,
				StartTime:      c.StartTime,
				EndTime:        c.EndTime,
			},
		}
		order.Connections = append(order.Connections, conn)
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	data, err := order.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableOrder)
	key = append(key, block.Previous.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	key = key[0:0]
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, param.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)

	data, err = ctx.GetStorage(nil, key)
	if err != nil {
		userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
		userInfo.ProductIds = make([]*abi.DoDSettleProduct, 0)
		userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

		internalId := &abi.DoDSettleInternalIdWrap{InternalId: block.Previous}
		userInfo.InternalIds = append(userInfo.InternalIds, internalId)
	} else {
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		internalId := &abi.DoDSettleInternalIdWrap{InternalId: block.Previous}
		userInfo.InternalIds = append(userInfo.InternalIds, internalId)
	}

	data, err = userInfo.MarshalMsg(nil)
	if err != nil {
		return err
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}

func (co *DoDSettleChangeOrder) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(abi.DoDSettleResponseParam)
	_, err := param.UnmarshalMsg(block.GetPayload())
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, input.Previous)
	if err != nil {
		return nil, err
	}

	if param.Action == abi.DoDSettleResponseActionConfirm {
		order.ContractState = abi.DoDSettleContractStateConfirmed
	} else {
		order.ContractState = abi.DoDSettleContractStateRejected
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = order.Seller.Address
	block.Token = cfg.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.NewBalance(0)
	block.Oracle = types.NewBalance(0)
	block.Storage = types.NewBalance(0)
	block.Network = types.NewBalance(0)

	am, _ := ctx.GetAccountMeta(block.Address)
	if am != nil {
		tm := am.Token(cfg.GasToken())
		if tm != nil {
			block.Balance = tm.Balance
			block.Representative = tm.Representative
			block.Previous = tm.Header
		} else {
			block.Balance = types.NewBalance(0)
			if len(am.Tokens) > 0 {
				block.Representative = am.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = types.NewBalance(0)
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	err = abi.DoDSettleUpdateOrder(ctx, order, input.Previous)
	if err != nil {
		return nil, err
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: order.Seller.Address,
			BlockType: types.ContractReward,
			Amount:    types.NewBalance(0),
			Token:     cfg.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (co *DoDSettleChangeOrder) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	param := new(abi.DoDSettleChangeOrderParam)
	err := param.FromABI(block.GetPayload())
	if err != nil {
		return types.ZeroAddress, err
	}

	return param.Seller.Address, nil
}

type DoDSettleTerminateOrder struct {
	BaseContract
}

func (to *DoDSettleTerminateOrder) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	_, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken())
	if err != nil {
		return nil, nil, err
	}

	param := new(abi.DoDSettleTerminateOrderParam)
	err = param.FromABI(block.GetPayload())
	if err != nil {
		return nil, nil, err
	}

	err = param.Verify()
	if err != nil {
		return nil, nil, err
	}

	if param.Buyer.Address != block.Address {
		return nil, nil, fmt.Errorf("invalid operator")
	}

	err = to.setStorage(ctx, param, block)
	if err != nil {
		return nil, nil, err
	}

	return &types.PendingKey{
			Address: param.Seller.Address,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: param.Buyer.Address,
			Amount: types.ZeroBalance,
			Type:   block.Token,
		}, nil
}

func (to *DoDSettleTerminateOrder) setStorage(ctx *vmstore.VMContext, param *abi.DoDSettleTerminateOrderParam, block *types.StateBlock) error {
	order := abi.NewOrderInfo()

	order.Buyer.Address = param.Buyer.Address
	order.Buyer.Name = param.Buyer.Name
	order.Seller.Address = param.Seller.Address
	order.Seller.Name = param.Seller.Name
	order.OrderType = abi.DoDSettleOrderTypeTerminate
	order.ContractState = abi.DoDSettleContractStateRequest

	for _, pid := range param.ProductId {
		conn := &abi.DoDSettleConnectionParam{
			DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
				ProductId: pid,
			},
		}
		order.Connections = append(order.Connections, conn)
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	data, err := order.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableOrder)
	key = append(key, block.Previous.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	key = key[0:0]
	key = append(key, abi.DoDSettleDBTableUser)
	key = append(key, param.Buyer.Address.Bytes()...)

	userInfo := new(abi.DoDSettleUserInfos)

	data, err = ctx.GetStorage(nil, key)
	if err != nil {
		userInfo.InternalIds = make([]*abi.DoDSettleInternalIdWrap, 0)
		userInfo.ProductIds = make([]*abi.DoDSettleProduct, 0)
		userInfo.OrderIds = make([]*abi.DoDSettleOrder, 0)

		internalId := &abi.DoDSettleInternalIdWrap{InternalId: block.Previous}
		userInfo.InternalIds = append(userInfo.InternalIds, internalId)
	} else {
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		internalId := &abi.DoDSettleInternalIdWrap{InternalId: block.Previous}
		userInfo.InternalIds = append(userInfo.InternalIds, internalId)
	}

	data, err = userInfo.MarshalMsg(nil)
	if err != nil {
		return err
	}

	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}

func (to *DoDSettleTerminateOrder) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(abi.DoDSettleResponseParam)
	_, err := param.UnmarshalMsg(block.GetPayload())
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, input.Previous)
	if err != nil {
		return nil, err
	}

	if param.Action == abi.DoDSettleResponseActionConfirm {
		order.ContractState = abi.DoDSettleContractStateConfirmed
	} else {
		order.ContractState = abi.DoDSettleContractStateRejected
	}

	// generate contract reward block
	block.Type = types.ContractReward
	block.Address = order.Seller.Address
	block.Token = cfg.GasToken()
	block.Link = input.GetHash()
	block.PoVHeight = input.PoVHeight
	block.Timestamp = common.TimeNow().Unix()

	// pledge fields only for QLC token
	block.Vote = types.NewBalance(0)
	block.Oracle = types.NewBalance(0)
	block.Storage = types.NewBalance(0)
	block.Network = types.NewBalance(0)

	am, _ := ctx.GetAccountMeta(block.Address)
	if am != nil {
		tm := am.Token(cfg.GasToken())
		if tm != nil {
			block.Balance = tm.Balance
			block.Representative = tm.Representative
			block.Previous = tm.Header
		} else {
			block.Balance = types.NewBalance(0)
			if len(am.Tokens) > 0 {
				block.Representative = am.Tokens[0].Representative
			} else {
				block.Representative = input.Representative
			}
			block.Previous = types.ZeroHash
		}
	} else {
		block.Balance = types.NewBalance(0)
		block.Representative = input.Representative
		block.Previous = types.ZeroHash
	}

	track := &abi.DoDSettleOrderLifeTrack{
		ContractState: order.ContractState,
		Time:          block.Timestamp,
		Hash:          block.Previous,
	}
	order.Track = append(order.Track, track)

	err = abi.DoDSettleUpdateOrder(ctx, order, input.Previous)
	if err != nil {
		return nil, err
	}

	return []*ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: order.Seller.Address,
			BlockType: types.ContractReward,
			Amount:    types.NewBalance(0),
			Token:     cfg.GasToken(),
			Data:      []byte{},
		},
	}, nil
}

func (to *DoDSettleTerminateOrder) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	param := new(abi.DoDSettleTerminateOrderParam)
	err := param.FromABI(block.GetPayload())
	if err != nil {
		return types.ZeroAddress, err
	}

	return param.Seller.Address, nil
}

type DoDSettleResourceReady struct {
	BaseContract
}

func (rr *DoDSettleResourceReady) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.GasToken() {
		return nil, nil, ErrToken
	}

	// make sure the gas account is activated
	_, err := ctx.GetTokenMeta(block.GetAddress(), cfg.GasToken())
	if err != nil {
		return nil, nil, err
	}

	param := new(abi.DoDSettleResourceReadyParam)
	err = param.FromABI(block.GetPayload())
	if err != nil {
		return nil, nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByOrderId(ctx, block.Address, param.OrderId)
	if err != nil {
		return nil, nil, err
	}

	if order.Seller.Address != block.Address {
		return nil, nil, fmt.Errorf("invalid operator")
	}

	if order.OrderState == abi.DoDSettleOrderStateFail {
		return nil, nil, fmt.Errorf("invalid order state")
	}

	for _, pid := range param.ProductId {
		var conn *abi.DoDSettleConnectionInfo
		var connParam *abi.DoDSettleConnectionParam

		productKey := &abi.DoDSettleProduct{
			Seller:    order.Seller.Address,
			ProductId: pid,
		}
		productHash := productKey.Hash()

		act := &abi.DoDSettleConnectionActive{ActiveAt: time.Now().Unix()}
		err = abi.DodSettleSetSellerConnectionActive(ctx, act, productHash)
		if err != nil {
			return nil, nil, err
		}

		for _, c := range order.Connections {
			if c.ProductId == pid {
				connParam = c
				break
			}
		}

		if connParam == nil {
			return nil, nil, fmt.Errorf("illegal operation")
		}

		// only update payg
		if connParam.BillingType == abi.DoDSettleBillingTypeDOD {
			continue
		}

		if order.OrderType == abi.DoDSettleOrderTypeCreate {
			conn, _ = abi.DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
			if conn != nil {
				return nil, nil, fmt.Errorf("dup product")
			}

			conn = &abi.DoDSettleConnectionInfo{
				DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
					ProductId:      pid,
					SrcCompanyName: connParam.SrcCompanyName,
					SrcRegion:      connParam.SrcRegion,
					SrcCity:        connParam.SrcCity,
					SrcDataCenter:  connParam.SrcDataCenter,
					SrcPort:        connParam.SrcPort,
					DstCompanyName: connParam.DstCompanyName,
					DstRegion:      connParam.DstRegion,
					DstCity:        connParam.DstCity,
					DstDataCenter:  connParam.DstDataCenter,
					DstPort:        connParam.DstPort,
				},
				Active: &abi.DoDSettleConnectionDynamicParam{
					ConnectionName: connParam.ConnectionName,
					PaymentType:    connParam.PaymentType,
					BillingType:    connParam.BillingType,
					Currency:       connParam.Currency,
					ServiceClass:   connParam.ServiceClass,
					Bandwidth:      connParam.Bandwidth,
					BillingUnit:    connParam.BillingUnit,
					Price:          connParam.Price,
					StartTime:      time.Now().Unix(),
				},
				Done:  make([]*abi.DoDSettleConnectionDynamicParam, 0),
				Track: make([]*abi.DoDSettleConnectionLifeTrack, 0),
			}
		} else if order.OrderType == abi.DoDSettleOrderTypeChange {
			conn, err = abi.DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
			if err != nil {
				return nil, nil, err
			}

			newActive := &abi.DoDSettleConnectionDynamicParam{
				ConnectionName: connParam.ConnectionName,
				PaymentType:    connParam.PaymentType,
				BillingType:    connParam.BillingType,
				Currency:       connParam.Currency,
				ServiceClass:   connParam.ServiceClass,
				Bandwidth:      connParam.Bandwidth,
				BillingUnit:    connParam.BillingUnit,
				Price:          connParam.Price,
				StartTime:      connParam.StartTime,
				EndTime:        connParam.EndTime,
			}

			abi.DoDSettleInheritParam(conn.Active, newActive)

			conn.Active.EndTime = abi.DoDSettleBillingUnitRound(conn.Active.BillingUnit, conn.Active.StartTime, time.Now().Unix())
			newActive.StartTime = conn.Active.EndTime

			conn.Done = append(conn.Done, conn.Active)
			conn.Active = newActive
		} else {
			conn, err = abi.DoDSettleGetConnectionInfoByProductHash(ctx, productHash)
			if err != nil {
				return nil, nil, err
			}

			conn.Active.EndTime = time.Now().Unix()
			conn.Done = append(conn.Done, conn.Active)
			conn.Active = nil
		}

		track := &abi.DoDSettleConnectionLifeTrack{
			OrderType: order.OrderType,
			OrderId:   order.OrderId,
			Time:      time.Now().Unix(),
		}

		if order.OrderType == abi.DoDSettleOrderTypeCreate || order.OrderType == abi.DoDSettleOrderTypeChange {
			track.Changed = &abi.DoDSettleConnectionDynamicParam{
				ConnectionName: connParam.ConnectionName,
				PaymentType:    connParam.PaymentType,
				BillingType:    connParam.BillingType,
				Currency:       connParam.Currency,
				ServiceClass:   connParam.ServiceClass,
				Bandwidth:      connParam.Bandwidth,
				BillingUnit:    connParam.BillingUnit,
				Price:          connParam.Price,
				StartTime:      connParam.StartTime,
				EndTime:        connParam.EndTime,
			}
		}

		conn.Track = append(conn.Track, track)

		err = abi.DoDSettleUpdateConnection(ctx, conn, productHash)
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, nil, nil
}
