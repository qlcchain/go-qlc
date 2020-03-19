package api

import (
	"errors"
	"fmt"
	"sort"

	"go.uber.org/zap"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/log"
	"github.com/qlcchain/go-qlc/vm/contract"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

type PermissionApi struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	cc     *chainctx.ChainContext
	ctx    *vmstore.VMContext
	au     *contract.AdminUpdate
	na     *contract.NodeAdd
	nu     *contract.NodeUpdate
	nr     *contract.NodeRemove
}

func NewPermissionApi(cfgFile string, l ledger.Store) *PermissionApi {
	api := &PermissionApi{
		l:      l,
		logger: log.NewLogger("api permission"),
		cc:     chainctx.NewChainContext(cfgFile),
		ctx:    vmstore.NewVMContext(l),
		au:     &contract.AdminUpdate{},
		na:     &contract.NodeAdd{},
		nu:     &contract.NodeUpdate{},
		nr:     &contract.NodeRemove{},
	}
	return api
}

type AdminUpdateParam struct {
	Admin     types.Address `json:"admin"`
	Successor types.Address `json:"successor"`
	Comment   string        `json:"comment"`
}

type AdminUser struct {
	Account types.Address `json:"admin"`
	Status  string        `json:"status"`
	Comment string        `json:"comment"`
}

type NodeParam struct {
	Admin   types.Address `json:"admin"`
	Index   uint32        `json:"index"`
	Kind    uint8         `json:"kind"`
	Node    string        `json:"node"`
	Comment string        `json:"comment"`
}

type NodeInfo struct {
	Index   uint32 `json:"index"`
	Kind    uint8  `json:"kind"`
	Node    string `json:"node"`
	Comment string `json:"comment"`
}

func (p *PermissionApi) GetAdminUpdateSendBlock(param *AdminUpdateParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if len(param.Comment) > abi.PermissionCommentMaxLen {
		return nil, fmt.Errorf("invalid comment len")
	}

	oldAdmin, err := abi.GetPermissionAdmin(p.ctx)
	if err != nil {
		return nil, fmt.Errorf("get admin err")
	}

	if oldAdmin.Addr != param.Admin || oldAdmin.Status != abi.PermissionAdminStatusActive {
		return nil, fmt.Errorf("invalid admin")
	}

	data, err := abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, param.Successor, param.Comment)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := p.l.GetAccountMeta(param.Admin)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Admin)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Admin,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(types.PermissionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.au.SetStorage(vmContext, oldAdmin)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PermissionApi) GetAdminUpdateRewardBlock(input *types.StateBlock) (*types.StateBlock, error) {
	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if input.GetType() != types.ContractSend {
		return nil, errors.New("input block type is not contract send")
	}

	if input.GetLink() != types.PermissionAddress.ToHash() {
		return nil, errors.New("input address is not contract permission")
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	reward := &types.StateBlock{}
	reward.PoVHeight = povHeader.GetHeight()

	vmContext := vmstore.NewVMContext(p.l)
	blocks, err := p.au.DoReceive(vmContext, reward, input)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		reward.Extra = *h
	}

	if len(blocks) > 0 {
		return reward, nil
	}

	return nil, errors.New("can not generate reward recv block")
}

func (p *PermissionApi) GetAdminUpdateRewardBlockBySendHash(sendHash types.Hash) (*types.StateBlock, error) {
	input, err := p.l.GetStateBlockConfirmed(sendHash)
	if err != nil {
		return nil, err
	}

	return p.GetAdminUpdateRewardBlock(input)
}

func (p *PermissionApi) GetAdmin() (*AdminUser, error) {
	a, err := abi.GetPermissionAdmin(p.ctx)
	if err != nil {
		return nil, err
	}

	au := &AdminUser{
		Account: a.Addr,
		Status:  abi.PermissionAdminStatusString(a.Status),
		Comment: a.Comment,
	}
	return au, nil
}

func (p *PermissionApi) GetNodeAddBlock(param *NodeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	pn := &abi.PermNode{
		Kind:    param.Kind,
		Node:    param.Node,
		Comment: param.Comment,
	}

	err := p.na.VerifyParam(p.ctx, pn)
	if err != nil {
		return nil, err
	}

	admin, err := abi.GetPermissionAdmin(p.ctx)
	if err != nil {
		return nil, errors.New("get admin err")
	}

	if admin.Addr != param.Admin || admin.Status != abi.PermissionAdminStatusActive {
		return nil, errors.New("invalid admin")
	}

	data, err := abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, param.Kind, param.Node, param.Comment)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := p.l.GetAccountMeta(param.Admin)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Admin)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Admin,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(types.PermissionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.na.SetStorage(vmContext, pn)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PermissionApi) GetNodeUpdateBlock(param *NodeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	pn := &abi.PermNode{
		Index:   param.Index,
		Kind:    param.Kind,
		Node:    param.Node,
		Comment: param.Comment,
	}

	err := p.nu.VerifyParam(p.ctx, pn)
	if err != nil {
		return nil, err
	}

	admin, err := abi.GetPermissionAdmin(p.ctx)
	if err != nil {
		return nil, errors.New("get admin err")
	}

	if admin.Addr != param.Admin || admin.Status != abi.PermissionAdminStatusActive {
		return nil, errors.New("invalid admin")
	}

	data, err := abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, param.Index, param.Kind, param.Node, param.Comment)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := p.l.GetAccountMeta(param.Admin)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Admin)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Admin,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(types.PermissionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.nu.SetStorage(vmContext, pn)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PermissionApi) GetNodeRemoveBlock(param *NodeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	pn := &abi.PermNode{
		Index: param.Index,
	}

	err := p.nr.VerifyParam(p.ctx, pn)
	if err != nil {
		return nil, err
	}

	admin, err := abi.GetPermissionAdmin(p.ctx)
	if err != nil {
		return nil, errors.New("get admin err")
	}

	if admin.Addr != param.Admin || admin.Status != abi.PermissionAdminStatusActive {
		return nil, errors.New("invalid admin")
	}

	data, err := abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeRemove, param.Index)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	am, err := p.l.GetAccountMeta(param.Admin)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Admin)
	}

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Admin,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           am.CoinVote,
		Network:        am.CoinNetwork,
		Oracle:         am.CoinOracle,
		Storage:        am.CoinStorage,
		Link:           types.Hash(types.PermissionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	vmContext := vmstore.NewVMContext(p.l)
	err = p.nr.SetStorage(vmContext, pn)
	if err != nil {
		return nil, err
	}

	h := vmContext.Cache.Trie().Hash()
	if h != nil {
		send.Extra = *h
	}

	return send, nil
}

func (p *PermissionApi) GetNodeByIndex(index uint32) (*NodeInfo, error) {
	n, err := abi.GetPermissionNode(p.ctx, index)
	if err != nil {
		return nil, err
	}

	if n.Valid == false {
		return nil, errors.New("node has been removed")
	}

	ni := &NodeInfo{
		Index:   n.Index,
		Kind:    n.Kind,
		Node:    n.Node,
		Comment: n.Comment,
	}
	return ni, nil
}

func (p *PermissionApi) GetNodesCount() int {
	ns, err := abi.GetAllPermissionNodes(p.ctx)
	if err != nil {
		return 0
	} else {
		return len(ns)
	}
}

func (p *PermissionApi) GetNodes(count int, offset int) ([]*NodeInfo, error) {
	ns, err := abi.GetAllPermissionNodes(p.ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(ns, func(i, j int) bool {
		return ns[i].Index < ns[j].Index
	})

	i := 0
	nis := make([]*NodeInfo, 0)
	for _, n := range ns {
		if n.Valid == false {
			continue
		}

		i++
		if i <= offset {
			continue
		}

		if i > offset+count {
			break
		}

		ni := &NodeInfo{
			Index:   n.Index,
			Kind:    n.Kind,
			Node:    n.Node,
			Comment: n.Comment,
		}
		nis = append(nis, ni)
	}

	return nis, nil
}
