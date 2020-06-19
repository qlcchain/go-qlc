package apis

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	chainctx "github.com/qlcchain/go-qlc/chain/context"
	"github.com/qlcchain/go-qlc/common/statedb"
	"github.com/qlcchain/go-qlc/common/topic"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/vmcontract/contractaddress"
	"github.com/qlcchain/go-qlc/ledger"
	"github.com/qlcchain/go-qlc/mock"
	"github.com/qlcchain/go-qlc/rpc/api"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"github.com/qlcchain/go-qlc/vm/contract/abi"
)

func addTestAdmin(t *testing.T, l *ledger.Ledger, admin *abi.AdminAccount, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	if err != nil {
		t.Fatal(err)
	}

	trieKey := statedb.PovCreateContractLocalStateKey(abi.PermissionDataAdmin, admin.Account.Bytes())

	data, err := admin.MarshalMsg(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = csdb.SetValue(trieKey, data)
	if err != nil {
		t.Fatal(err)
	}

	err = gsdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}
	txn := l.DBStore().Batch(true)
	err = gsdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	povBlk.Header.CbTx.StateHash = gsdb.GetCurHash()
	mock.UpdatePovHash(povBlk)

	err = l.AddPovBlock(povBlk, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.SetPovLatestHeight(povBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
}

func addTestNodes(t *testing.T, l *ledger.Ledger, pns []*abi.PermNode, povHeight uint64) {
	povBlk, povTd := mock.GeneratePovBlockByFakePow(nil, 0)
	povBlk.Header.BasHdr.Height = povHeight

	gsdb := statedb.NewPovGlobalStateDB(l.DBStore(), types.ZeroHash)
	csdb, err := gsdb.LookupContractStateDB(contractaddress.PermissionAddress)
	if err != nil {
		t.Fatal(err)
	}

	for _, pn := range pns {
		trieKey := statedb.PovCreateContractLocalStateKey(abi.PermissionDataNode, []byte(pn.NodeId))

		data, err := pn.MarshalMsg(nil)
		if err != nil {
			t.Fatal(err)
		}

		err = csdb.SetValue(trieKey, data)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = gsdb.CommitToTrie()
	if err != nil {
		t.Fatal(err)
	}
	txn := l.DBStore().Batch(true)
	err = gsdb.CommitToDB(txn)
	if err != nil {
		t.Fatal(err)
	}
	err = l.DBStore().PutBatch(txn)
	if err != nil {
		t.Fatal(err)
	}

	povBlk.Header.CbTx.StateHash = gsdb.GetCurHash()
	mock.UpdatePovHash(povBlk)

	err = l.AddPovBlock(povBlk, povTd)
	if err != nil {
		t.Fatal(err)
	}

	err = l.AddPovBestHash(povBlk.GetHeight(), povBlk.GetHash())
	if err != nil {
		t.Fatal(err)
	}

	err = l.SetPovLatestHeight(povBlk.GetHeight())
	if err != nil {
		t.Fatal(err)
	}
}

func TestPermissionApi_GetAdminUpdateBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionAPI(cfgFile, l)
	param := new(api.AdminUpdateParam)
	param.Comment = strings.Repeat("x", abi.PermissionCommentMaxLen+1)

	_, err := p.GetAdminHandoverBlock(context.Background(), &pb.AdminUpdateParam{
		Admin:     toAddressValue(param.Admin),
		Successor: toAddressValue(param.Successor),
		Comment:   param.Comment,
	})
	if err != chainctx.ErrPoVNotFinish {
		t.Fatal()
	}
}

func TestPermissionApi_GetAdmin(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionAPI(cfgFile, l)
	a, err := p.GetAdmin(context.Background(), nil)
	if err == nil {
		t.Fatal()
	}

	ac := &abi.AdminAccount{
		Account: mock.Address(),
		Comment: "test",
		Valid:   true,
	}
	addTestAdmin(t, l, ac, 10)

	a, err = p.GetAdmin(context.Background(), nil)
	if err != nil || a.Account != ac.Account.String() || a.Comment != ac.Comment {
		t.Fatal()
	}
}

func TestPermissionApi_GetNodeUpdateBlock(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()
	cc := chainctx.NewChainContext(cfgFile)
	p := NewPermissionAPI(cfgFile, l)
	param := new(api.NodeParam)

	blk, err := p.GetNodeUpdateBlock(context.Background(), &pb.NodeParam{
		Admin:   toAddressValue(param.Admin),
		NodeId:  param.NodeId,
		NodeUrl: param.NodeUrl,
		Comment: param.Comment,
	})
	if err != chainctx.ErrPoVNotFinish {
		t.Fatal()
	}

	cc.Init(nil)
	cc.EventBus().Publish(topic.EventPovSyncState, topic.SyncDone)
	time.Sleep(time.Second)
	blk, _ = p.GetNodeUpdateBlock(context.Background(), &pb.NodeParam{
		Admin:   toAddressValue(param.Admin),
		NodeId:  param.NodeId,
		NodeUrl: param.NodeUrl,
		Comment: param.Comment,
	})
	if blk != nil {
		t.Fatal()
	}

	param.NodeId = "QmVLbouTEb9LGQJ56KvQCyoPXqDeqwYSE6j1YSyfLeHgN3"
	param.NodeUrl = "123"
	param.Comment = "test"
	blk, _ = p.GetNodeUpdateBlock(context.Background(), &pb.NodeParam{
		Admin:   toAddressValue(param.Admin),
		NodeId:  param.NodeId,
		NodeUrl: param.NodeUrl,
		Comment: param.Comment,
	})
	if blk != nil {
		t.Fatal()
	}
}

func TestPermissionApi_GetNode(t *testing.T) {
	clear, l, cfgFile := getTestLedger()
	if l == nil {
		t.Fatal()
	}
	defer clear()

	p := NewPermissionAPI(cfgFile, l)
	nodeId := "n1"

	_, err := p.GetNode(context.Background(), toString(nodeId))
	if err == nil {
		t.Fatal()
	}

	pn := &abi.PermNode{NodeId: nodeId, NodeUrl: "", Comment: "", Valid: true}
	addTestNodes(t, l, []*abi.PermNode{pn}, 10)
	_, err = p.GetNode(context.Background(), toString(nodeId))
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

	p := NewPermissionAPI(cfgFile, l)

	count, err := p.GetNodesCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if count.GetValue() != 0 {
		t.Fatal()
	}

	pns := make([]*abi.PermNode, 0)
	for i := 0; i < 15; i++ {
		pn := &abi.PermNode{NodeId: fmt.Sprintf("n%d", i), Valid: true}
		pns = append(pns, pn)
	}

	addTestNodes(t, l, pns, 1)

	count, err = p.GetNodesCount(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if count.GetValue() != 15 {
		t.Fatal()
	}

	ns, err := p.GetNodes(context.Background(), &pb.Offset{
		Count:  5,
		Offset: 6,
	})
	if err != nil {
		t.Fatal()
	}

	if len(ns.GetNodes()) != 5 {
		t.Fatal()
	}
}
