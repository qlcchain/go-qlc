package api

import (
	"strings"
	"testing"
	"time"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func addTestNode(t *testing.T, ctx *vmstore.VMContext, pn *abi.PermNode) {
	data, err := pn.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.PermissionDataNode)
	key = append(key, util.BE_Uint32ToBytes(pn.Index)...)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}
}

func addTestAdmin(t *testing.T, ctx *vmstore.VMContext, admin *abi.AdminAccount) {
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.PermissionDataAdmin)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		t.Fatal()
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetAdminUpdateSendBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	param := new(AdminUpdateParam)
	param.Comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)

	blk, err := p.GetAdminUpdateSendBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	blk, err = p.GetAdminUpdateSendBlock(param)
	if err != chainctx.ErrPoVNotFinish {
		t.Fatal()
	}

	p.cc.Init(nil)
	p.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Comment = strings.Repeat("x", abi.PermissionCommentMaxLen)
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk != nil {
		t.Fatal()
	}

	abi.PermissionInit(p.ctx)
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Admin = cfg.GenesisAddress()
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk != nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Admin)
	l.AddAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am.Tokens[0].Type = cfg.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetAdminUpdateSendBlock(param)
	if blk == nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetAdminUpdateRewardBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	send := mock.StateBlockWithoutWork()

	blk, _ := p.GetAdminUpdateRewardBlock(send)
	if blk != nil {
		t.Fatal()
	}

	p.cc.Init(nil)
	p.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = p.GetAdminUpdateRewardBlock(send)
	if blk != nil {
		t.Fatal()
	}

	send.Type = types.ContractSend
	blk, _ = p.GetAdminUpdateRewardBlock(send)
	if blk != nil {
		t.Fatal()
	}

	send.Link = contractaddress.PermissionAddress.ToHash()
	blk, _ = p.GetAdminUpdateRewardBlock(send)
	if blk != nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	blk, _ = p.GetAdminUpdateRewardBlock(send)
	if blk != nil {
		t.Fatal()
	}

	send.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, mock.Address(), "successor")
	blk, _ = p.GetAdminUpdateRewardBlock(send)
	if blk == nil {
		t.Fatal()
	}

	t.Log(blk)
}

func TestPermissionApi_GetAdminUpdateRewardBlockBySendHash(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	send := mock.StateBlockWithoutWork()
	send.Type = types.ContractSend
	send.Link = contractaddress.PermissionAddress.ToHash()
	send.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, mock.Address(), "successor")
	p.cc.Init(nil)
	p.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())

	blk, _ := p.GetAdminUpdateRewardBlockBySendHash(send.GetHash())
	if blk != nil {
		t.Fatal()
	}

	l.AddStateBlock(send)
	blk, _ = p.GetAdminUpdateRewardBlockBySendHash(send.GetHash())
	if blk == nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetAdmin(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	a, err := p.GetAdmin()
	if err == nil {
		t.Fatal()
	}

	ac := &abi.AdminAccount{
		Addr:    mock.Address(),
		Comment: "test",
		Status:  abi.PermissionAdminStatusActive,
	}
	addTestAdmin(t, p.ctx, ac)

	a, err = p.GetAdmin()
	if err != nil || a.Account != ac.Addr || a.Status != abi.PermissionAdminStatusString(ac.Status) || a.Comment != ac.Comment {
		t.Fatal()
	}
}

func TestPermissionApi_GetNodeAddBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	param := new(NodeParam)

	blk, err := p.GetNodeAddBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	blk, err = p.GetNodeAddBlock(param)
	if err != chainctx.ErrPoVNotFinish {
		t.Fatal()
	}

	p.cc.Init(nil)
	p.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = p.GetNodeAddBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Kind = abi.PermissionNodeKindPeerID
	param.Node = "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3"
	param.Comment = "test"
	blk, _ = p.GetNodeAddBlock(param)
	if blk != nil {
		t.Fatal()
	}

	abi.PermissionInit(p.ctx)
	blk, _ = p.GetNodeAddBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Admin = cfg.GenesisAddress()
	blk, _ = p.GetNodeAddBlock(param)
	if blk != nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	blk, _ = p.GetNodeAddBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Admin)
	l.AddAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetNodeAddBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am.Tokens[0].Type = cfg.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetNodeAddBlock(param)
	if blk == nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetNodeUpdateBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	param := new(NodeParam)

	blk, err := p.GetNodeUpdateBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	blk, err = p.GetNodeUpdateBlock(param)
	if err != chainctx.ErrPoVNotFinish {
		t.Fatal()
	}

	p.cc.Init(nil)
	p.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Index = 1
	param.Kind = abi.PermissionNodeKindPeerID
	param.Node = "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3"
	param.Comment = "test"
	pn := &abi.PermNode{
		Index:   1,
		Kind:    param.Kind,
		Node:    param.Node,
		Comment: param.Comment,
		Valid:   true,
	}
	addTestNode(t, p.ctx, pn)
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk != nil {
		t.Fatal()
	}

	abi.PermissionInit(p.ctx)
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Admin = cfg.GenesisAddress()
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk != nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Admin)
	l.AddAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am.Tokens[0].Type = cfg.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetNodeUpdateBlock(param)
	if blk == nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetNodeRemoveBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	param := new(NodeParam)

	blk, err := p.GetNodeRemoveBlock(nil)
	if err != ErrParameterNil {
		t.Fatal()
	}

	blk, err = p.GetNodeRemoveBlock(param)
	if err != chainctx.ErrPoVNotFinish {
		t.Fatal()
	}

	p.cc.Init(nil)
	p.cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Index = 1
	param.Kind = abi.PermissionNodeKindPeerID
	param.Node = "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3"
	param.Comment = "test"
	pn := &abi.PermNode{
		Index:   1,
		Kind:    param.Kind,
		Node:    param.Node,
		Comment: param.Comment,
		Valid:   true,
	}
	addTestNode(t, p.ctx, pn)
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk != nil {
		t.Fatal()
	}

	abi.PermissionInit(p.ctx)
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk != nil {
		t.Fatal()
	}

	param.Admin = cfg.GenesisAddress()
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk != nil {
		t.Fatal()
	}

	pb, td := mock.GeneratePovBlock(nil, 0)
	l.AddPovBlock(pb, td)
	l.SetPovLatestHeight(pb.Header.BasHdr.Height)
	l.AddPovBestHash(pb.Header.BasHdr.Height, pb.GetHash())
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(param.Admin)
	l.AddAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk != nil {
		t.Fatal()
	}

	am.Tokens[0].Type = cfg.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	blk, _ = p.GetNodeRemoveBlock(param)
	if blk == nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetNodeByIndex(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)
	index := uint32(10)
	_, err := p.GetNodeByIndex(index)
	if err == nil {
		t.Fatal()
	}

	pn := &abi.PermNode{Index: index, Valid: true}
	addTestNode(t, p.ctx, pn)
	_, err = p.GetNodeByIndex(index)
	if err != nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetNodes(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionApi(cfgFile, l)

	count := p.GetNodesCount()
	if count != 0 {
		t.Fatal()
	}

	i := uint32(0)
	for ; i < 15; i++ {
		pn := &abi.PermNode{Index: i, Valid: true}
		addTestNode(t, p.ctx, pn)
	}

	count = p.GetNodesCount()
	if count != 15 {
		t.Fatal()
	}

	ns, err := p.GetNodes(5, 6)
	if err != nil {
		t.Fatal()
	}

	if len(ns) != 5 || ns[0].Index != 6 || ns[4].Index != 10 {
		t.Fatal()
	}
}
