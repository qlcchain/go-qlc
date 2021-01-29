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

type PermissionApi struct {
	logger *zap.SugaredLogger
	l      ledger.Store
	cc     *chainctx.ChainContext
	ctx    *vmstore.VMContext
	au     *contract.AdminHandOver
	nu     *contract.NodeUpdate
}

func NewPermissionApi(cfgFile string, l ledger.Store) *PermissionApi {
	api := &PermissionApi{
		l:      l,
		logger: log.NewLogger("api permission"),
		cc:     chainctx.NewChainContext(cfgFile),
		ctx:    vmstore.NewVMContext(l, &contractaddress.PermissionAddress),
		au:     &contract.AdminHandOver{},
		nu:     &contract.NodeUpdate{},
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
	Comment string        `json:"comment"`
}

type NodeParam struct {
	Admin   types.Address `json:"admin"`
	NodeId  string        `json:"nodeId"`
	NodeUrl string        `json:"nodeUrl"`
	Comment string        `json:"comment"`
}

type NodeInfo struct {
	NodeId  string `json:"nodeId"`
	NodeUrl string `json:"nodeUrl"`
	Comment string `json:"comment"`
}

func (p *PermissionApi) GetAdminHandoverBlock(param *AdminUpdateParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	if len(param.Comment) > abi.PermissionCommentMaxLen {
		return nil, fmt.Errorf("invalid comment len")
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	if !abi.PermissionIsAdmin(p.ctx, param.Admin) {
		return nil, fmt.Errorf("user %s is not admin", param.Admin)
	}

	am, err := p.l.GetAccountMeta(param.Admin)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Admin)
	}

	data, err := abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminHandOver, param.Successor, param.Comment)

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Admin,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Link:           types.Hash(contractaddress.PermissionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	return send, nil
}

func (p *PermissionApi) GetAdmin() (*AdminUser, error) {
	a, err := abi.PermissionGetAdmin(p.ctx)
	if err != nil {
		return nil, err
	}

	au := &AdminUser{
		Account: a[0].Account,
		Comment: a[0].Comment,
	}
	return au, nil
}

func (p *PermissionApi) GetNodeUpdateBlock(param *NodeParam) (*types.StateBlock, error) {
	if param == nil {
		return nil, ErrParameterNil
	}

	if !p.cc.IsPoVDone() {
		return nil, chainctx.ErrPoVNotFinish
	}

	pn := &abi.PermNode{
		NodeId:  param.NodeId,
		NodeUrl: param.NodeUrl,
		Comment: param.Comment,
	}

	err := p.nu.VerifyParam(p.ctx, pn)
	if err != nil {
		return nil, err
	}

	povHeader, err := p.l.GetLatestPovHeader()
	if err != nil {
		return nil, fmt.Errorf("get pov header error: %s", err)
	}

	if !abi.PermissionIsAdmin(p.ctx, param.Admin) {
		return nil, fmt.Errorf("user %s is not admin", param.Admin)
	}

	am, err := p.l.GetAccountMeta(param.Admin)
	if err != nil {
		return nil, err
	}

	tm := am.Token(config.ChainToken())
	if tm == nil {
		return nil, fmt.Errorf("%s do not have qlc token", param.Admin)
	}

	data, err := abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, param.NodeId, param.NodeUrl, param.Comment)

	send := &types.StateBlock{
		Type:           types.ContractSend,
		Token:          tm.Type,
		Address:        param.Admin,
		Balance:        tm.Balance,
		Previous:       tm.Header,
		Vote:           types.ToBalance(am.CoinVote),
		Network:        types.ToBalance(am.CoinNetwork),
		Oracle:         types.ToBalance(am.CoinOracle),
		Storage:        types.ToBalance(am.CoinStorage),
		Link:           types.Hash(contractaddress.PermissionAddress),
		Representative: tm.Representative,
		Data:           data,
		PoVHeight:      povHeader.GetHeight(),
		Timestamp:      common.TimeNow().Unix(),
	}

	return send, nil
}

func (p *PermissionApi) GetNode(id string) (*NodeInfo, error) {
	n, err := abi.PermissionGetNode(p.ctx, id)
	if err != nil {
		return nil, err
	}

	ni := &NodeInfo{
		NodeId:  n.NodeId,
		NodeUrl: n.NodeUrl,
		Comment: n.Comment,
	}
	return ni, nil
}

func (p *PermissionApi) GetNodesCount() int {
	ns, err := abi.PermissionGetAllNodes(p.l)
	if err != nil {
		return 0
	} else {
		return len(ns)
	}
}

func (p *PermissionApi) GetNodes(count int, offset int) ([]*NodeInfo, error) {
	ns, err := abi.PermissionGetAllNodes(p.l)
	if err != nil {
		return nil, err
	}

	nis := make([]*NodeInfo, 0)
	for i, n := range ns {
		if i < offset {
			continue
		}

		if i >= offset+count {
			break
		}

		ni := &NodeInfo{
			NodeId:  n.NodeId,
			NodeUrl: n.NodeUrl,
			Comment: n.Comment,
		}
		nis = append(nis, ni)
	}

	return nis, nil
}
