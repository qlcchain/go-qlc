package contract

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/event"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type AdminUpdate struct {
	BaseContract
}

func (a *AdminUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.AdminAccount)
	err := abi.PermissionABI.UnpackMethod(admin, abi.MethodNamePermissionAdminUpdate, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	if len(admin.Comment) > abi.PermissionCommentMaxLen {
		return nil, nil, ErrInvalidLen
	}

	oldAdmin, err := abi.GetPermissionAdmin(ctx)
	if err != nil {
		return nil, nil, ErrGetAdmin
	}

	if oldAdmin.Addr != block.Address || oldAdmin.Status != abi.PermissionAdminStatusActive {
		return nil, nil, ErrInvalidAdmin
	}

	block.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, admin.Addr, admin.Comment)

	oldAdmin.Status = abi.PermissionAdminStatusHandOver
	err = a.SetStorage(ctx, oldAdmin)
	if err != nil {
		return nil, nil, ErrSetStorage
	}

	return &types.PendingKey{
			Address: admin.Addr,
			Hash:    block.GetHash(),
		}, &types.PendingInfo{
			Source: types.Address(block.Link),
			Type:   cfg.ChainToken(),
		}, nil
}

func (a *AdminUpdate) SetStorage(ctx *vmstore.VMContext, admin *abi.AdminAccount) error {
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PermissionDataAdmin)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (a *AdminUpdate) DoReceive(ctx *vmstore.VMContext, block *types.StateBlock, input *types.StateBlock) ([]*vmcontract.ContractBlock, error) {
	admin := new(abi.AdminAccount)
	err := abi.PermissionABI.UnpackMethod(admin, abi.MethodNamePermissionAdminUpdate, input.Data)
	if err != nil {
		return nil, ErrUnpackMethod
	}

	am, _ := ctx.Ledger.GetAccountMeta(admin.Addr)
	if am != nil {
		tm := am.Token(cfg.ChainToken())
		block.Balance = am.CoinBalance
		block.Vote = am.CoinVote
		block.Network = am.CoinNetwork
		block.Oracle = am.CoinOracle
		block.Storage = am.CoinStorage
		if tm != nil {
			block.Previous = tm.Header
			block.Representative = tm.Representative
		} else {
			block.Previous = types.ZeroHash
			block.Representative = input.Representative
		}
	} else {
		block.Vote = types.ZeroBalance
		block.Network = types.ZeroBalance
		block.Oracle = types.ZeroBalance
		block.Storage = types.ZeroBalance
		block.Previous = types.ZeroHash
		block.Representative = input.Representative
	}

	block.Type = types.ContractReward
	block.Address = admin.Addr
	block.Token = cfg.ChainToken()
	block.Link = input.GetHash()
	block.Timestamp = common.TimeNow().Unix()

	admin.Status = abi.PermissionAdminStatusActive
	err = a.SetStorage(ctx, admin)
	if err != nil {
		return nil, ErrSetStorage
	}

	return []*vmcontract.ContractBlock{
		{
			VMContext: ctx,
			Block:     block,
			ToAddress: admin.Addr,
			BlockType: types.ContractReward,
			Amount:    types.NewBalance(0),
			Token:     cfg.ChainToken(),
			Data:      []byte{},
		},
	}, nil
}

func (a *AdminUpdate) GetTargetReceiver(ctx *vmstore.VMContext, block *types.StateBlock) (types.Address, error) {
	admin := new(abi.AdminAccount)
	err := abi.PermissionABI.UnpackMethod(admin, abi.MethodNamePermissionAdminUpdate, block.Data)
	if err != nil {
		return types.ZeroAddress, err
	} else {
		return admin.Addr, nil
	}
}

type NodeAdd struct {
	BaseContract
}

func (n *NodeAdd) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeAdd, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = n.VerifyParam(ctx, node)
	if err != nil {
		return nil, nil, ErrCheckParam
	}

	admin, err := abi.GetPermissionAdmin(ctx)
	if err != nil {
		return nil, nil, ErrGetAdmin
	}

	if admin.Addr != block.Address || admin.Status != abi.PermissionAdminStatusActive {
		return nil, nil, ErrInvalidAdmin
	}

	block.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, node.Kind, node.Node, node.Comment)

	err = n.SetStorage(ctx, node)
	if err != nil {
		return nil, nil, ErrSetStorage
	}

	return nil, nil, nil
}

func (n *NodeAdd) VerifyParam(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	switch pn.Kind {
	// x.x.x.x:xx
	case abi.PermissionNodeKindIPPort:
		s := strings.Split(pn.Node, ":")
		if len(s) != 2 {
			return errors.New("node format err")
		}

		ip := net.ParseIP(s[0])
		if ip == nil {
			return errors.New("invalid node ip")
		}

		port, err := strconv.Atoi(s[1])
		if err != nil || port <= 0 || port > 65535 {
			return errors.New("invalid node port")
		}
	case abi.PermissionNodeKindPeerID:
		if len(pn.Node) != 46 {
			return errors.New("invalid node peer id")
		}
	default:
		return errors.New("invalid node type")
	}

	if len(pn.Comment) > abi.PermissionCommentMaxLen {
		return fmt.Errorf("invalid node comment len (max %d)", abi.PermissionCommentMaxLen)
	}

	return nil
}

func (n *NodeAdd) SetStorage(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	index := abi.GetPermissionNodeIndex(ctx)

	pn.Valid = true
	data, err := pn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PermissionDataNode)
	key = append(key, util.BE_Uint32ToBytes(index)...)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		return err
	}

	key = key[0:0]
	key = append(key, abi.PermissionDataNodeIndex)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, util.BE_Uint32ToBytes(index+1))
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeAdd) EventNotify(eb event.EventBus, ctx *vmstore.VMContext, block *types.StateBlock) error {
	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeAdd, block.Data)
	if err != nil {
		return ErrUnpackMethod
	}

	index := abi.GetPermissionNodeIndex(ctx)
	pe := &common.PermissionEvent{
		EventType: common.PermissionEventNodeAdd,
		NodeKind:  node.Kind,
		NodeAddr:  node.Node,
		NodeIndex: index - 1,
	}
	eb.Publish(topic.EventPermissionNodeUpdate, pe)

	return nil
}

type NodeUpdate struct {
	BaseContract
}

func (n *NodeUpdate) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeUpdate, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = n.VerifyParam(ctx, node)
	if err != nil {
		return nil, nil, ErrCheckParam
	}

	admin, err := abi.GetPermissionAdmin(ctx)
	if err != nil {
		return nil, nil, ErrGetAdmin
	}

	if admin.Addr != block.Address || admin.Status != abi.PermissionAdminStatusActive {
		return nil, nil, ErrInvalidAdmin
	}

	block.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, node.Index, node.Kind, node.Node, node.Comment)

	err = n.SetStorage(ctx, node)
	if err != nil {
		return nil, nil, ErrSetStorage
	}

	return nil, nil, nil
}

func (n *NodeUpdate) VerifyParam(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	node, err := abi.GetPermissionNode(ctx, pn.Index)
	if err != nil || node.Valid == false {
		return err
	}

	switch pn.Kind {
	// x.x.x.x:xx
	case abi.PermissionNodeKindIPPort:
		s := strings.Split(pn.Node, ":")
		if len(s) != 2 {
			return errors.New("node format err")
		}

		ip := net.ParseIP(s[0])
		if ip == nil {
			return errors.New("invalid node ip")
		}

		port, err := strconv.Atoi(s[1])
		if err != nil || port <= 0 || port > 65535 {
			return errors.New("invalid node port")
		}
	case abi.PermissionNodeKindPeerID:
		if len(pn.Node) != 46 {
			return errors.New("invalid node peer id")
		}
	default:
		return errors.New("invalid node type")
	}

	if len(pn.Comment) > abi.PermissionCommentMaxLen {
		return fmt.Errorf("invalid node comment len (max %d)", abi.PermissionCommentMaxLen)
	}

	return nil
}

func (n *NodeUpdate) SetStorage(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	pn.Valid = true
	data, err := pn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PermissionDataNode)
	key = append(key, util.BE_Uint32ToBytes(pn.Index)...)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeUpdate) EventNotify(eb event.EventBus, ctx *vmstore.VMContext, block *types.StateBlock) error {
	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeUpdate, block.Data)
	if err != nil {
		return ErrUnpackMethod
	}

	pe := &common.PermissionEvent{
		EventType: common.PermissionEventNodeUpdate,
		NodeKind:  node.Kind,
		NodeAddr:  node.Node,
		NodeIndex: node.Index,
	}
	eb.Publish(topic.EventPermissionNodeUpdate, pe)

	return nil
}

type NodeRemove struct {
	BaseContract
}

func (n *NodeRemove) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeRemove, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	err = n.VerifyParam(ctx, node)
	if err != nil {
		return nil, nil, ErrCheckParam
	}

	admin, err := abi.GetPermissionAdmin(ctx)
	if err != nil {
		return nil, nil, ErrGetAdmin
	}

	if admin.Addr != block.Address || admin.Status != abi.PermissionAdminStatusActive {
		return nil, nil, ErrInvalidAdmin
	}

	block.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeRemove, node.Index)

	err = n.SetStorage(ctx, node)
	if err != nil {
		return nil, nil, ErrSetStorage
	}

	return nil, nil, nil
}

func (n *NodeRemove) VerifyParam(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	node, err := abi.GetPermissionNode(ctx, pn.Index)
	if err != nil || node.Valid == false {
		return err
	}

	return nil
}

func (n *NodeRemove) SetStorage(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	pn.Valid = false
	data, err := pn.MarshalMsg(nil)
	if err != nil {
		return err
	}

	var key []byte
	key = append(key, abi.PermissionDataNode)
	key = append(key, util.BE_Uint32ToBytes(pn.Index)...)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		return err
	}

	return nil
}

func (n *NodeRemove) EventNotify(eb event.EventBus, ctx *vmstore.VMContext, block *types.StateBlock) error {
	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeRemove, block.Data)
	if err != nil {
		return ErrUnpackMethod
	}

	pe := &common.PermissionEvent{
		EventType: common.PermissionEventNodeRemove,
		NodeIndex: node.Index,
	}
	eb.Publish(topic.EventPermissionNodeUpdate, pe)

	return nil
}
