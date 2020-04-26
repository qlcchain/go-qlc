package contract

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

var PermissionContract = vmcontract.NewChainContract(
	map[string]vmcontract.Contract{
		abi.MethodNamePermissionAdminHandOver: &AdminHandOver{
			BaseContract: BaseContract{
				Describe: vmcontract.Describe{
					SpecVer:   vmcontract.SpecVer2,
					Signature: true,
					Work:      true,
					PovState:  true,
				},
			},
		},
		abi.MethodNamePermissionNodeUpdate: &NodeUpdate{
			BaseContract: BaseContract{
				Describe: vmcontract.Describe{
					SpecVer:   vmcontract.SpecVer2,
					Signature: true,
					Work:      true,
					PovState:  true,
				},
			},
		},
	},
	abi.PermissionABI,
	abi.JsonPermission,
)

type AdminHandOver struct {
	BaseContract
}

func (a *AdminHandOver) ProcessSend(ctx *vmstore.VMContext, block *types.StateBlock) (*types.PendingKey, *types.PendingInfo, error) {
	if block.GetToken() != cfg.ChainToken() {
		return nil, nil, ErrToken
	}

	admin := new(abi.AdminAccount)
	err := abi.PermissionABI.UnpackMethod(admin, abi.MethodNamePermissionAdminHandOver, block.Data)
	if err != nil {
		return nil, nil, ErrUnpackMethod
	}

	if len(admin.Comment) > abi.PermissionCommentMaxLen {
		return nil, nil, ErrInvalidLen
	}

	// check admin if the block is not synced
	if !block.IsFromSync() {
		if !abi.PermissionIsAdmin(ctx, block.Address) {
			return nil, nil, ErrInvalidAdmin
		}
	}

	block.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminHandOver, admin.Account, admin.Comment)

	return nil, nil, nil
}

func (a *AdminHandOver) removeAdmin(csdb *statedb.PovContractStateDB, addr types.Address) error {
	trieKey := statedb.PovCreateContractLocalStateKey(abi.PermissionDataAdmin, addr.Bytes())

	// genesis admin is not in db
	if addr == cfg.GenesisAddress() {
		return nil
	}

	valBytes, err := csdb.GetValue(trieKey)
	if err != nil || len(valBytes) == 0 {
		return errors.New("get admin err")
	}

	admin := new(abi.AdminAccount)
	_, err = admin.UnmarshalMsg(valBytes)
	if err != nil {
		return err
	}

	admin.Valid = false
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}

func (a *AdminHandOver) updateAdmin(csdb *statedb.PovContractStateDB, addr types.Address, comment string) error {
	trieKey := statedb.PovCreateContractLocalStateKey(abi.PermissionDataAdmin, addr.Bytes())

	// genesis admin is not in db
	if addr == cfg.GenesisAddress() {
		return nil
	}

	admin := new(abi.AdminAccount)
	admin.Comment = comment
	admin.Valid = true
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}

func (a *AdminHandOver) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	admin := new(abi.AdminAccount)
	err := abi.PermissionABI.UnpackMethod(admin, abi.MethodNamePermissionAdminHandOver, block.Data)
	if err != nil {
		return err
	}

	err = a.removeAdmin(csdb, block.Address)
	if err != nil {
		return err
	}

	return a.updateAdmin(csdb, admin.Account, admin.Comment)
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

	// check admin if the block is not synced
	if !block.IsFromSync() {
		if !abi.PermissionIsAdmin(ctx, block.Address) {
			return nil, nil, ErrInvalidAdmin
		}
	}

	block.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, node.NodeId, node.NodeUrl, node.Comment)

	return nil, nil, nil
}

func (n *NodeUpdate) VerifyParam(ctx *vmstore.VMContext, pn *abi.PermNode) error {
	if len(pn.NodeUrl) > 0 {
		s := strings.Split(pn.NodeUrl, ":")
		if len(s) > 2 {
			return errors.New("url format err")
		}

		ip := net.ParseIP(s[0])
		if ip == nil {
			return errors.New("invalid url ip")
		}

		if len(s) == 2 {
			port, err := strconv.Atoi(s[1])
			if err != nil || port <= 0 || port > 65535 {
				return errors.New("invalid url port")
			}
		}
	}

	if len(pn.Comment) > abi.PermissionCommentMaxLen {
		return fmt.Errorf("invalid node comment len (max %d)", abi.PermissionCommentMaxLen)
	}

	return nil
}

func (n *NodeUpdate) DoSendOnPov(ctx *vmstore.VMContext, csdb *statedb.PovContractStateDB, povHeight uint64, block *types.StateBlock) error {
	node := new(abi.PermNode)
	err := abi.PermissionABI.UnpackMethod(node, abi.MethodNamePermissionNodeUpdate, block.Data)
	if err != nil {
		return err
	}

	pe := &topic.PermissionEvent{
		EventType: topic.PermissionEventNodeUpdate,
		NodeId:    node.NodeId,
		NodeUrl:   node.NodeUrl,
	}
	ctx.EventBus().Publish(topic.EventPermissionNodeUpdate, pe)

	trieKey := statedb.PovCreateContractLocalStateKey(abi.PermissionDataNode, []byte(node.NodeId))

	node.Valid = true
	data, err := node.MarshalMsg(nil)
	if err != nil {
		return err
	}

	return csdb.SetValue(trieKey, data)
}
