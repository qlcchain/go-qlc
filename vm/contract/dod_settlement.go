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
	err = param.FromABI(block.Data)
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
	order.State = abi.DoDSettleOrderStateCreateRequest

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
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
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
		userInfo.Infos = make([]*abi.DoDSettleUserInfo, 0)
		info := &abi.DoDSettleUserInfo{InternalId: block.Previous}
		userInfo.Infos = append(userInfo.Infos, info)
	} else {
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		info := &abi.DoDSettleUserInfo{InternalId: block.Previous}
		userInfo.Infos = append(userInfo.Infos, info)
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
	_, err := param.UnmarshalMsg(block.Data)
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, input.Previous)
	if err != nil {
		return nil, err
	}

	if param.Action == abi.DoDResponseActionConfirm {
		order.State = abi.DoDSettleOrderStateCreateConfirmed
	} else {
		order.State = abi.DoDSettleOrderStateRejected
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
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
	}
	order.Track = append(order.Track, track)

	err = co.updateOrder(ctx, order, input.Previous)
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

func (co *DoDSettleCreateOrder) updateOrder(ctx *vmstore.VMContext, order *abi.DoDSettleOrderInfo, id types.Hash) error {
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

	return nil
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
	err = param.FromABI(block.Data)
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

	switch param.Operation {
	case abi.DoDOrderOperationCreate:
		for i, c := range order.Connections {
			productIdHash := types.Sha256DHashData([]byte(param.ProductId[i]))

			conn, err := abi.DoDSettleGetConnectionInfoByProductId(ctx, productIdHash)
			if conn != nil {
				return nil, nil, fmt.Errorf("dup product")
			}

			order.Connections[i].ProductId = param.ProductId[i]

			conn = &abi.DoDSettleConnectionInfo{
				OrderId: param.OrderId,
				DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
					ProductId:      param.ProductId[i],
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
				Pending: &abi.DoDSettleConnectionDynamicParam{
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
				Active: nil,
				Done:   make([]*abi.DoDSettleConnectionDynamicParam, 0),
				Track:  make([]*abi.DoDSettleConnectionLifeTrack, 0),
			}

			err = uo.updateConnection(ctx, conn, productIdHash)
			if err != nil {
				return nil, nil, err
			}
		}

		order.State = abi.DoDSettleOrderStateCreateSend
	case abi.DoDOrderOperationChange:
		for _, c := range order.Connections {
			productIdHash := types.Sha256DHashData([]byte(c.ProductId))

			conn, err := abi.DoDSettleGetConnectionInfoByProductId(ctx, productIdHash)
			if err != nil {
				return nil, nil, fmt.Errorf("dup product")
			}

			conn.OrderId = param.OrderId
			conn.Ready = false
			conn.Pending = &abi.DoDSettleConnectionDynamicParam{
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
			}

			uo.inheritParam(conn.Pending, conn.Active)

			err = uo.updateConnection(ctx, conn, productIdHash)
			if err != nil {
				return nil, nil, err
			}
		}

		order.State = abi.DoDSettleOrderStateChangeSend
	case abi.DoDOrderOperationTerminate:
		for _, c := range order.Connections {
			productIdHash := types.Sha256DHashData([]byte(c.ProductId))

			conn, err := abi.DoDSettleGetConnectionInfoByProductId(ctx, productIdHash)
			if err != nil {
				return nil, nil, fmt.Errorf("dup product")
			}

			conn.OrderId = param.OrderId
			conn.Ready = false
			conn.Pending = nil

			err = uo.updateConnection(ctx, conn, productIdHash)
			if err != nil {
				return nil, nil, err
			}
		}

		order.State = abi.DoDSettleOrderStateTerminateSend
	case abi.DoDOrderOperationFail:
		order.State = abi.DoDSettleOrderStateFailed
	}

	order.OrderId = param.OrderId

	track := &abi.DoDSettleOrderLifeTrack{
		State:  order.State,
		Reason: param.FailReason,
		Time:   block.Timestamp,
		Hash:   block.Previous,
	}
	order.Track = append(order.Track, track)

	err = uo.updateOrder(ctx, order, param.InternalId)
	if err != nil {
		return nil, nil, err
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

func (uo *DoDSettleUpdateOrderInfo) inheritParam(src, dst *abi.DoDSettleConnectionDynamicParam) {
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

func (uo *DoDSettleUpdateOrderInfo) updateConnection(ctx *vmstore.VMContext, conn *abi.DoDSettleConnectionInfo, id types.Hash) error {
	data, err := conn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableProduct)
	key = append(key, id.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}

func (uo *DoDSettleUpdateOrderInfo) DoGap(ctx *vmstore.VMContext, block *types.StateBlock) (common.ContractGapType, interface{}, error) {
	param := new(abi.DoDSettleUpdateOrderInfoParam)
	err := param.FromABI(block.Data)
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	if param.InternalId.IsZero() {
		return common.ContractNoGap, nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, param.InternalId)
	if err != nil {
		return common.ContractNoGap, nil, err
	}

	switch param.Operation {
	case abi.DoDOrderOperationCreate:
		if order.State != abi.DoDSettleOrderStateCreateConfirmed {
			return common.ContractDoDOrderState, param.InternalId, nil
		}
	case abi.DoDOrderOperationChange:
		if order.State != abi.DoDSettleOrderStateChangeConfirmed {
			return common.ContractDoDOrderState, param.InternalId, nil
		}
	case abi.DoDOrderOperationTerminate:
		if order.State != abi.DoDSettleOrderStateTerminateConfirmed {
			return common.ContractDoDOrderState, param.InternalId, nil
		}
	case abi.DoDOrderOperationFail:
		if order.State != abi.DoDSettleOrderStateCreateConfirmed && order.State != abi.DoDSettleOrderStateChangeConfirmed &&
			order.State != abi.DoDSettleOrderStateTerminateConfirmed {
			return common.ContractDoDOrderState, param.InternalId, nil
		}
	}

	return common.ContractNoGap, nil, nil
}

func (uo *DoDSettleUpdateOrderInfo) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*ContractBlock, error) {
	param := new(abi.DoDSettleUpdateOrderInfoParam)
	err := param.FromABI(input.Data)
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, param.InternalId)
	if err != nil {
		return nil, err
	}

	order.State = abi.DoDSettleOrderStateComplete

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
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
	}
	order.Track = append(order.Track, track)

	err = uo.updateOrder(ctx, order, param.InternalId)
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

func (uo *DoDSettleUpdateOrderInfo) updateOrder(ctx *vmstore.VMContext, order *abi.DoDSettleOrderInfo, id types.Hash) error {
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

	return nil
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
	err = param.FromABI(block.Data)
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
	order.State = abi.DoDSettleOrderStateChangeRequest

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
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
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
		userInfo.Infos = make([]*abi.DoDSettleUserInfo, 0)
		info := &abi.DoDSettleUserInfo{InternalId: block.Previous}
		userInfo.Infos = append(userInfo.Infos, info)
	} else {
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		info := &abi.DoDSettleUserInfo{InternalId: block.Previous}
		userInfo.Infos = append(userInfo.Infos, info)
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
	_, err := param.UnmarshalMsg(block.Data)
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, input.Previous)
	if err != nil {
		return nil, err
	}

	if param.Action == abi.DoDResponseActionConfirm {
		order.State = abi.DoDSettleOrderStateChangeConfirmed
	} else {
		order.State = abi.DoDSettleOrderStateRejected
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
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
	}
	order.Track = append(order.Track, track)

	err = co.updateOrder(ctx, order, input.Previous)
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

func (co *DoDSettleChangeOrder) updateOrder(ctx *vmstore.VMContext, order *abi.DoDSettleOrderInfo, id types.Hash) error {
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

	return nil
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
	err = param.FromABI(block.Data)
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
	order.State = abi.DoDSettleOrderStateTerminateRequest

	for _, pid := range param.ProductId {
		conn := &abi.DoDSettleConnectionParam{
			DoDSettleConnectionStaticParam: abi.DoDSettleConnectionStaticParam{
				ProductId: pid,
			},
		}
		order.Connections = append(order.Connections, conn)
	}

	track := &abi.DoDSettleOrderLifeTrack{
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
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
		userInfo.Infos = make([]*abi.DoDSettleUserInfo, 0)
		info := &abi.DoDSettleUserInfo{InternalId: block.Previous}
		userInfo.Infos = append(userInfo.Infos, info)
	} else {
		_, err = userInfo.UnmarshalMsg(data)
		if err != nil {
			return err
		}

		info := &abi.DoDSettleUserInfo{InternalId: block.Previous}
		userInfo.Infos = append(userInfo.Infos, info)
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
	_, err := param.UnmarshalMsg(block.Data)
	if err != nil {
		return nil, err
	}

	order, err := abi.DoDSettleGetOrderInfoByInternalId(ctx, input.Previous)
	if err != nil {
		return nil, err
	}

	if param.Action == abi.DoDResponseActionConfirm {
		order.State = abi.DoDSettleOrderStateTerminateConfirmed
	} else {
		order.State = abi.DoDSettleOrderStateRejected
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
		State: order.State,
		Time:  block.Timestamp,
		Hash:  block.Previous,
	}
	order.Track = append(order.Track, track)

	err = to.updateOrder(ctx, order, input.Previous)
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

func (to *DoDSettleTerminateOrder) updateOrder(ctx *vmstore.VMContext, order *abi.DoDSettleOrderInfo, id types.Hash) error {
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

	return nil
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
	err = param.FromABI(block.Data)
	if err != nil {
		return nil, nil, err
	}

	for _, pid := range param.ProductId {
		productIdHash := types.Sha256DHashData([]byte(pid))

		conn, err := abi.DoDSettleGetConnectionInfoByProductId(ctx, productIdHash)
		if err != nil {
			return nil, nil, err
		}

		if conn.OrderId != param.OrderId || conn.Ready {
			return nil, nil, err
		}

		conn.Ready = true

		// if order.Seller.Address != block.Address {
		// 	return nil, nil, fmt.Errorf("invalid operator")
		// }

		if conn.Active != nil {
			if conn.Active.BillingType == abi.DoDBillingTypePAYG {
				conn.Active.EndTime = time.Now().Unix()
			}

			conn.Done = append(conn.Done, conn.Active)
			conn.Active = conn.Pending
			conn.Pending = nil

			if conn.Active != nil && conn.Active.BillingType == abi.DoDBillingTypePAYG {
				conn.Active.StartTime = time.Now().Unix()
			}
		} else {
			if conn.Pending == nil {
				return nil, nil, fmt.Errorf("null connection info")
			}

			conn.Active = conn.Pending
			conn.Pending = nil

			if conn.Active.BillingType == abi.DoDBillingTypePAYG {
				conn.Active.StartTime = time.Now().Unix()
			}
		}

		err = rr.setStorage(ctx, conn, productIdHash)
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, nil, nil
}

func (rr *DoDSettleResourceReady) setStorage(ctx *vmstore.VMContext, conn *abi.DoDSettleConnectionInfo, id types.Hash) error {
	data, err := conn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.DoDSettleDBTableProduct)
	key = append(key, id.Bytes()...)
	err = ctx.SetStorage(nil, key, data)
	if err != nil {
		return err
	}

	return nil
}
