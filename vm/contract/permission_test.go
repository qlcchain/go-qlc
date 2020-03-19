package contract

import (
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/common/event"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func addTestAdmin(t *testing.T, ctx *vmstore.VMContext, admin *abi.AdminAccount) {
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.PermissionDataAdmin)
	err = ctx.SetStorage(types.PermissionAddress[:], key, data)
	if err != nil {
		t.Fatal()
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal()
	}
}

func addTestNode(t *testing.T, ctx *vmstore.VMContext, pn *abi.PermNode) {
	data, err := pn.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, abi.PermissionDataNode)
	key = append(key, util.BE_Uint32ToBytes(pn.Index)...)
	err = ctx.SetStorage(types.PermissionAddress[:], key, data)
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdminUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	a := new(AdminUpdate)

	blk.Token = mock.Hash()
	_, _, err := a.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	newAdminAddr := mock.Address()
	newAdminComment := strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, newAdminAddr, newAdminComment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidLen {
		t.Fatal(err)
	}

	newAdminComment = strings.Repeat("x", abi.PermissionCommentMaxLen)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, newAdminAddr, newAdminComment)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrGetAdmin {
		t.Fatal(err)
	}

	admin := &abi.AdminAccount{
		Addr:    mock.Address(),
		Comment: "old Admin",
		Status:  abi.PermissionAdminStatusHandOver,
	}
	addTestAdmin(t, ctx, admin)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	blk.Address = admin.Addr
	_, _, err = a.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	admin.Status = abi.PermissionAdminStatusActive
	addTestAdmin(t, ctx, admin)
	_, _, err = a.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAdminUpdate_DoReceive(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	send := mock.StateBlockWithoutWork()
	recv := mock.StateBlockWithoutWork()
	a := new(AdminUpdate)

	_, err := a.DoReceive(ctx, recv, send)
	if err != ErrUnpackMethod {
		t.Fatal()
	}

	adminAddr := mock.Address()
	adminComment := "new Admin"
	send.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, adminAddr, adminComment)
	_, err = a.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	am := mock.AccountMeta(adminAddr)
	l.AddAccountMeta(am, l.Cache().GetCache())
	_, err = a.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}

	am.Tokens[0].Type = cfg.ChainToken()
	l.UpdateAccountMeta(am, l.Cache().GetCache())
	_, err = a.DoReceive(ctx, recv, send)
	if err != nil {
		t.Fatal()
	}
}

func TestAdminUpdate_GetTargetReceiver(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	a := new(AdminUpdate)

	r, _ := a.GetTargetReceiver(ctx, blk)
	if r != types.ZeroAddress {
		t.Fatal()
	}

	adminAddr := mock.Address()
	adminComment := strings.Repeat("x", abi.PermissionCommentMaxLen)
	blk.Data, _ = abi.PermissionABI.PackMethod(abi.MethodNamePermissionAdminUpdate, adminAddr, adminComment)
	r, _ = a.GetTargetReceiver(ctx, blk)
	if r != adminAddr {
		t.Fatal()
	}
}

func TestNodeAdd_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	n := new(NodeAdd)

	blk.Token = mock.Hash()
	_, _, err := n.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	kind := abi.PermissionNodeKindInvalid
	node := ""
	comment := ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = ""
	comment = ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = "1:9999999"
	comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = "1.1.1.1:9999999"
	comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = "1.1.1.1:9999"
	comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindPeerID
	node = strings.Repeat("x", 46+1)
	comment = ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindPeerID
	node = strings.Repeat("x", 46)
	comment = ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeAdd, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrGetAdmin {
		t.Fatal(err)
	}

	abi.PermissionInit(ctx)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	blk.Address = cfg.GenesisAddress()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	eb := event.GetEventBus("test")
	err = n.EventNotify(eb, ctx, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestNodeUpdate_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	n := new(NodeUpdate)

	blk.Token = mock.Hash()
	_, _, err := n.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	index := uint32(10)
	kind := abi.PermissionNodeKindInvalid
	node := ""
	comment := ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	tn := &abi.PermNode{
		Index:   index,
		Kind:    abi.PermissionNodeKindPeerID,
		Node:    "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3",
		Comment: "test node",
		Valid:   true,
	}
	addTestNode(t, ctx, tn)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = ""
	comment = ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = "1:999999"
	comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = "1.1.1.1:999999"
	comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindIPPort
	node = "1.1.1.1:9999"
	comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindPeerID
	node = strings.Repeat("x", 46+1)
	comment = ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	kind = abi.PermissionNodeKindPeerID
	node = strings.Repeat("x", 46)
	comment = ""
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeUpdate, index, kind, node, comment)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrGetAdmin {
		t.Fatal(err)
	}

	abi.PermissionInit(ctx)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	blk.Address = cfg.GenesisAddress()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	eb := event.GetEventBus("test")
	err = n.EventNotify(eb, ctx, blk)
	if err != nil {
		t.Fatal()
	}
}

func TestNodeRemove_ProcessSend(t *testing.T) {
	clear, l := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	ctx := vmstore.NewVMContext(l)
	blk := mock.StateBlockWithoutWork()
	n := new(NodeRemove)

	blk.Token = mock.Hash()
	_, _, err := n.ProcessSend(ctx, blk)
	if err != ErrToken {
		t.Fatal(err)
	}

	blk.Token = cfg.ChainToken()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrUnpackMethod {
		t.Fatal(err)
	}

	index := uint32(10)
	blk.Data, err = abi.PermissionABI.PackMethod(abi.MethodNamePermissionNodeRemove, index)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrCheckParam {
		t.Fatal(err)
	}

	tn := &abi.PermNode{
		Index:   index,
		Kind:    abi.PermissionNodeKindPeerID,
		Node:    "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3",
		Comment: "test node",
		Valid:   true,
	}
	addTestNode(t, ctx, tn)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrGetAdmin {
		t.Fatal(err)
	}

	abi.PermissionInit(ctx)
	_, _, err = n.ProcessSend(ctx, blk)
	if err != ErrInvalidAdmin {
		t.Fatal(err)
	}

	blk.Address = cfg.GenesisAddress()
	_, _, err = n.ProcessSend(ctx, blk)
	if err != nil {
		t.Fatal(err)
	}

	eb := event.GetEventBus("test")
	err = n.EventNotify(eb, ctx, blk)
	if err != nil {
		t.Fatal()
	}
}
