package abi

import (
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	cfg "github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/vm/abi"
	"github.com/qlcchain/go-qlc/vm/vmstore"
)

func addTestNode(t *testing.T, ctx *vmstore.VMContext, pn *PermNode) {
	data, err := pn.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, PermissionDataNode)
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

func addTestAdmin(t *testing.T, ctx *vmstore.VMContext, admin *AdminAccount) {
	data, err := admin.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	var key []byte
	key = append(key, PermissionDataAdmin)
	err = ctx.SetStorage(contractaddress.PermissionAddress[:], key, data)
	if err != nil {
		t.Fatal()
	}

	err = ctx.SaveStorage()
	if err != nil {
		t.Fatal()
	}
}

func TestPermissionABI(t *testing.T) {
	_, err := abi.JSONToABIContract(strings.NewReader(JsonPermission))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermissionInit(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	err := PermissionInit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	admin, err := GetPermissionAdmin(ctx)
	if err != nil || admin.Addr != cfg.GenesisAddress() || admin.Status != PermissionAdminStatusActive ||
		admin.Comment != "Initial Admin" {
		t.Fatal()
	}

	nAdmin := &AdminAccount{
		Addr:    mock.Address(),
		Comment: "new admin",
		Status:  PermissionAdminStatusActive,
	}
	addTestAdmin(t, ctx, nAdmin)

	err = PermissionInit(ctx)
	if err != nil {
		t.Fatal(err)
	}

	admin, err = GetPermissionAdmin(ctx)
	if err != nil || admin.Addr != nAdmin.Addr || admin.Status != PermissionAdminStatusActive ||
		admin.Comment != "new admin" {
		t.Fatal()
	}
}

func TestGetPermissionNodeIndex(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	index := GetPermissionNodeIndex(ctx)
	if index != 0 {
		t.Fatal()
	}

	var key []byte
	key = append(key, PermissionDataNodeIndex)
	err := ctx.SetStorage(contractaddress.PermissionAddress[:], key, util.BE_Uint32ToBytes(10))
	if err != nil {
		t.Fatal(err)
	}

	index = GetPermissionNodeIndex(ctx)
	if index != 10 {
		t.Fatal()
	}
}

func TestGetPermissionNode(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)
	pn, _ := GetPermissionNode(ctx, 10)
	if pn != nil {
		t.Fatal()
	}

	node := &PermNode{
		Index:   10,
		Kind:    PermissionNodeKindIPPort,
		Node:    "127.0.0.1:9735",
		Comment: "test",
		Valid:   true,
	}
	addTestNode(t, ctx, node)

	pn, err := GetPermissionNode(ctx, 10)
	if err != nil || pn.Index != node.Index || pn.Kind != node.Kind || pn.Node != node.Node ||
		pn.Comment != node.Comment || pn.Valid != node.Valid {
		t.Fatal()
	}
}

func TestGetAllPermissionNodes(t *testing.T) {
	teardownTestCase, l := setupLedgerForTestCase(t)
	defer teardownTestCase(t)

	ctx := vmstore.NewVMContext(l)

	node1 := &PermNode{
		Index:   10,
		Kind:    PermissionNodeKindIPPort,
		Node:    "127.0.0.1:9735",
		Comment: "test1",
		Valid:   true,
	}
	addTestNode(t, ctx, node1)

	node2 := &PermNode{
		Index:   11,
		Kind:    PermissionNodeKindPeerID,
		Node:    "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3",
		Comment: "test1",
		Valid:   false,
	}
	addTestNode(t, ctx, node2)

	pns, err := GetAllPermissionNodes(ctx)
	if err != nil {
		t.Fatal()
	}

	for _, n := range pns {
		switch n.Index {
		case 10:
			if n.Valid != node1.Valid || n.Kind != node1.Kind || n.Node != node1.Node || n.Comment != node1.Comment {
				t.Fatal()
			}
		case 11:
			if n.Valid != node2.Valid || n.Kind != node2.Kind || n.Node != node2.Node || n.Comment != node2.Comment {
				t.Fatal()
			}
		default:
			t.Fatal()
		}
	}
}
