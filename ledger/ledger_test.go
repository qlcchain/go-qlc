package ledger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/test/mock"
)

func setupTestCase(t *testing.T) (func(t *testing.T), *Ledger) {
	t.Parallel()

	dir := filepath.Join(config.QlcTestDataDir(), "ledger", uuid.New().String())
	_ = os.RemoveAll(dir)
	l := NewLedger(dir)

	return func(t *testing.T) {
		//err := l.Store.Erase()
		err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//CloseLedger()
		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}, l
}

//var bc, _ = mock.BlockChain()

func TestLedger_Instance1(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger1")
	l1 := NewLedger(dir)
	l2 := NewLedger(dir)
	t.Logf("l1:%v,l2:%v", l1, l2)
	defer func() {
		l1.Close()
		//l2.Close()
		_ = os.RemoveAll(dir)
	}()
	b := reflect.DeepEqual(l1, l2)
	if l1 == nil || l2 == nil || !b {
		t.Fatal("error")
	}
	//_ = os.RemoveAll(dir)
}

func TestLedger_Instance2(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger1")
	dir2 := filepath.Join(config.QlcTestDataDir(), "ledger2")
	l1 := NewLedger(dir)
	l2 := NewLedger(dir2)
	defer func() {
		l1.Close()
		l2.Close()
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(dir2)
	}()
	if l1 == nil || l2 == nil || reflect.DeepEqual(l1, l2) {
		t.Fatal("error")
	}
}

func TestGetTxn(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	txn := l.Store.NewTransaction(false)
	fmt.Println(txn)
	txn2, flag := l.getTxn(false, txn)

	if flag {
		t.Fatal("get txn flag error")
	}

	if txn != txn2 {
		t.Fatal("txn!=tnx2")
	}
}

func TestLedger_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	e, err := l.Empty()

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("is empty %s", strconv.FormatBool(e))
}

func TestLedgerSession_BatchUpdate(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	err := l.BatchUpdate(func(txn db.StoreTxn) error {
		blk := mock.StateBlockWithoutWork()
		if err := l.AddStateBlock(blk); err != nil {
			t.Fatal()
		}
		if err := l.AddStateBlock(mock.StateBlockWithoutWork()); err != nil {
			t.Fatal()
		}
		if ok, err := l.HasStateBlock(blk.GetHash()); err != nil || !ok {
			t.Fatal()
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func addStateBlock(t *testing.T, l *Ledger) *types.StateBlock {
	blk := mock.StateBlockWithoutWork()
	if err := l.AddStateBlock(blk); err != nil {
		t.Fatal(err)
	}
	return blk
}

func addSmartContractBlock(t *testing.T, l *Ledger) *types.SmartContractBlock {
	jsonBlock := `{
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "mcvnzY+zF5mVDjsvknvPfFgRToMQAVI4wivQGRZBwerbUIvfrKD6/suZJWiFVOI5sbTa98bpY9+1cUhE2T9yidxSCpvZ4kkBVBMfcL3OJIqG",
      "abiLength": 81,
      "abiHash": "79dab43dcc97205918b297c3aba6259e3ab1ed7d0779dc78eec6f57e5d6307ce"
    },
    "owner": "qlc_1nawsw4yatupd47p3scd5x5i3s9szbsggxbxmfy56f8jroyu945i5seu1cdd",
	"isUseStorage": false,
    "type": "SmartContract",
    "address": "qlc_3watpnwym9i43kbkt35yfp8xnqo7c9ujp3b6udajza71mspjfzpnpdgoydzn",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000007bb1fe",
    "signature": "d9d71c82eccdca0324e102c089b28c1430b0ae61f2af809e6134b289d5186b16cbcb6fcd4bfc1424fd34aa40e9bdd05069bc56d05fecf833470d80d047048a05"
  }`
	blk := new(types.SmartContractBlock)
	_ = json.Unmarshal([]byte(jsonBlock), blk)
	if err := l.AddSmartContractBlock(blk); err != nil {
		t.Log(err)
	}
	return blk
}

func TestLedger_AddBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addStateBlock(t, l)
}

func TestLedger_GetBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	blk, err := l.GetStateBlock(block.GetHash())
	t.Log("blk,", blk)
	if err != nil || blk == nil {
		t.Fatal(err)
	}
}

func TestLedger_GetSmartContrantBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addSmartContractBlock(t, l)
	blk, err := l.GetSmartContractBlock(block.GetHash())
	t.Log("blk,", blk)
	if err != nil || blk == nil {
		t.Fatal(err)
	}
}

func TestLedger_HasSmartContrantBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addSmartContractBlock(t, l)
	b, err := l.HasSmartContractBlock(block.GetHash())
	t.Log(b)
	if err != nil || !b {
		t.Fatal(err)
	}
}

func TestLedger_GetSmartContrantBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addSmartContractBlock(t, l)
	blk, err := l.GetSmartContractBlock(block.GetHash())
	t.Log("blk,", blk)
	if err != nil || blk == nil {
		t.Fatal(err)
	}
	n, err := l.CountSmartContractBlocks()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(n)
	err = l.GetSmartContractBlocks(func(block *types.SmartContractBlock) error {
		fmt.Println(block)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetAllBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addStateBlock(t, l)
	addStateBlock(t, l)
	err := l.GetStateBlocks(func(block *types.StateBlock) error {
		t.Log(block)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_DeleteBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	block := addStateBlock(t, l)
	if err := l.DeleteStateBlock(block.GetHash()); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_HasBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	r, err := l.HasStateBlock(block.GetHash())
	if err != nil || !r {
		t.Fatal(err)
	}
	t.Log("hasblock,", r)

}

func TestLedger_GetRandomBlock_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	b, err := l.GetRandomStateBlock()

	if err != ErrStoreEmpty {
		t.Fatal(err)
	}
	t.Log("block ,", b)
}

func addUncheckedBlock(t *testing.T, l *Ledger) (hash types.Hash, block *types.StateBlock, kind types.UncheckedKind) {
	block = mock.StateBlockWithoutWork()
	hash = block.GetPrevious()
	kind = types.UncheckedKindPrevious
	if err := l.AddUncheckedBlock(hash, block, kind, types.UnSynchronized); err != nil {
		t.Fatal(err)
	}
	return
}

func TestLedger_AddUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
}

func TestLedger_GetUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := addUncheckedBlock(t, l)

	if b, s, err := l.GetUncheckedBlock(parentHash, kind); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("unchecked,%s", b)
		t.Log(s)
	}

}

func TestLedger_CountUncheckedBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
	addUncheckedBlock(t, l)

	c, err := l.CountUncheckedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("unchecked count,", c)
}

func TestLedger_HasUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := addUncheckedBlock(t, l)
	r, err := l.HasUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has unchecked,", r)
}

func TestLedger_GetUncheckedBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addUncheckedBlock(t, l)
	addUncheckedBlock(t, l)

	err := l.WalkUncheckedBlocks(func(block types.Block, kind types.UncheckedKind) error {
		t.Log(kind, block)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_DeleteUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := addUncheckedBlock(t, l)
	err := l.DeleteUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
}

func addAccountMeta(t *testing.T, l *Ledger) *types.AccountMeta {

	ac := mock.Account()
	am := mock.AccountMeta(ac.Address())
	if err := l.AddAccountMeta(am); err != nil {
		t.Fatal()
	}
	return am
}

func TestLedger_AddAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addAccountMeta(t, l)
}

func TestLedger_GetAccountMeta_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := mock.Address()
	_, err := l.GetAccountMeta(address)
	if err != ErrAccountNotFound {
		t.Fatal(err)
	}
}

func TestLedger_GetAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	am := addAccountMeta(t, l)
	a, err := l.GetAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", a)
	for _, token := range a.Tokens {
		t.Log("token,", token)
	}
}

func TestLedger_HasAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	r, err := l.HasAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has account,", r)
}

func TestLedger_DeleteAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	err := l.DeleteAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdateAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)

	err := l.AddOrUpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}

}

func TestLedger_UpdateAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	am := addAccountMeta(t, l)
	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)

	err := l.AddOrUpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}
}

func addTokenMeta(t *testing.T, l *Ledger) *types.TokenMeta {
	tm := addAccountMeta(t, l)
	token := mock.TokenMeta(tm.Address)
	if err := l.AddTokenMeta(token.BelongTo, token); err != nil {
		t.Fatal(err)
	}
	return token
}

func TestLedger_AddTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addTokenMeta(t, l)
}

func TestLedger_GetTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	token, err := l.GetTokenMeta(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token,", token)
}

func TestLedger_AddOrUpdateTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	token2 := mock.TokenMeta(token.BelongTo)
	err := l.AddOrUpdateTokenMeta(token.BelongTo, token2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_UpdateTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	token := addTokenMeta(t, l)
	token2 := mock.TokenMeta(token.BelongTo)
	err := l.AddOrUpdateTokenMeta(token.BelongTo, token2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_DelTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	token := addTokenMeta(t, l)
	err := l.DeleteTokenMeta(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetAccountMetas(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	addAccountMeta(t, l)

	err := l.GetAccountMetas(func(am *types.AccountMeta) error {
		t.Log(am)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_CountAccountMetas(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	addAccountMeta(t, l)
	num, err := l.CountAccountMetas()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", num)
}

func TestLedger_HasTokenMeta_False(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	token2 := mock.TokenMeta(token.BelongTo)
	has, err := l.HasTokenMeta(token.BelongTo, token2.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", has)
}

func TestLedger_HasTokenMeta_True(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	token := addTokenMeta(t, l)
	r, err := l.HasTokenMeta(token.BelongTo, token.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", r)
}

func addRepresentationWeight(t *testing.T, l *Ledger) types.Address {
	address := mock.Address()
	i, _ := random.Intn(math.MaxInt16)
	amount := types.Balance{Int: big.NewInt(int64(i))}

	err := l.AddRepresentation(address, amount)
	if err != nil {
		t.Fatal(err)
	}
	return address
}

func TestLedger_AddRepresentationWeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addRepresentationWeight(t, l)
}

func TestLedger_SubRepresentationWeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := addRepresentationWeight(t, l)
	amount := types.Balance{Int: big.NewInt(int64(1000))}
	err := l.SubRepresentation(address, amount)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_GetRepresentation(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := addRepresentationWeight(t, l)
	a, err := l.GetRepresentation(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("amount,", a)
}

func TestLedger_GetRepresentations(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addRepresentationWeight(t, l)
	addRepresentationWeight(t, l)

	err := l.GetRepresentations(func(address types.Address, balance types.Balance) error {
		t.Log(address, balance)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func addPending(t *testing.T, l *Ledger) (pendingkey types.PendingKey, pendinginfo types.PendingInfo) {
	address := mock.Address()
	hash := mock.Hash()

	i, _ := random.Intn(math.MaxUint32)
	balance := types.Balance{Int: big.NewInt(int64(i))}
	pendinginfo = types.PendingInfo{
		Source: address,
		Amount: balance,
		Type:   mock.Hash(),
	}
	pendingkey = types.PendingKey{Address: address, Hash: hash}
	t.Log(pendinginfo)
	err := l.AddPending(pendingkey, &pendinginfo)
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestLedger_AddPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addPending(t, l)
}

func TestLedger_GetPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addPending(t, l)
	pendingkey, _ := addPending(t, l)
	p, err := l.GetPending(pendingkey)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pending,", p)
}

func TestLedger_DeletePending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, _ := addPending(t, l)

	err := l.DeletePending(pendingkey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.GetPending(pendingkey); err != nil && err != ErrPendingNotFound {
		t.Fatal(err)
	}
	t.Log("delete pending success")
}

func TestLedger_SearchPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address := mock.Address()
	for idx := 0; idx < 10; idx++ {
		hash := mock.Hash()
		i, _ := random.Intn(math.MaxUint32)
		balance := types.Balance{Int: big.NewInt(int64(i))}
		v := &types.PendingInfo{
			Source: address,
			Amount: balance,
			Type:   mock.Hash(),
		}
		k := &types.PendingKey{Address: address, Hash: hash}
		err := l.AddPending(*k, v)
		if err != nil {
			t.Fatal(err)
		}
		//t.Log(idx, util.ToString(k), util.ToString(v))
	}
	//t.Log("build cache done")

	counter := 0
	err := l.SearchPending(address, func(key *types.PendingKey, value *types.PendingInfo) error {
		t.Log(counter, util.ToString(key), util.ToString(value))
		counter++
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if counter != 10 {
		t.Fatal("invalid", counter)
	}
}

func addFrontier(t *testing.T, l *Ledger) *types.Frontier {
	frontier := new(types.Frontier)
	frontier.HeaderBlock = mock.Hash()
	frontier.OpenBlock = mock.Hash()
	if err := l.AddFrontier(frontier); err != nil {
		t.Fatal()
	}
	return frontier
}

func TestLedger_AddFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addFrontier(t, l)
}

func TestLedger_GetFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	f := addFrontier(t, l)
	f, err := l.GetFrontier(f.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier,", f)
}

func TestLedger_GetAllFrontiers(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addFrontier(t, l)
	addFrontier(t, l)
	addFrontier(t, l)

	c, err := l.CountFrontiers()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier count,", c)

	fs, err := l.GetFrontiers()
	if err != nil {
		t.Fatal(err)
	}
	for index, f := range fs {
		t.Log("frontier", index, f)
	}
}

func TestLedger_DeleteFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	f := addFrontier(t, l)
	err := l.DeleteFrontier(f.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.GetFrontier(f.HeaderBlock); err != nil && err != ErrFrontierNotFound {
		t.Fatal(err)
	}
	t.Log("delete frontier success")
}

func TestReleaseLedger(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger1")
	dir2 := filepath.Join(config.QlcTestDataDir(), "ledger2")
	l1 := NewLedger(dir)
	_ = NewLedger(dir2)
	defer func() {
		//only release ledger1
		l1.Close()
		CloseLedger()
		_ = os.RemoveAll(dir)
		_ = os.RemoveAll(dir2)
	}()
}

func TestLedgerSession_Latest(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	token := mock.TokenMeta(block.GetAddress())
	token.Header = block.GetHash()
	token.Type = block.GetToken()
	ac := types.AccountMeta{Address: token.BelongTo, Tokens: []*types.TokenMeta{token}}
	if err := l.AddAccountMeta(&ac); err != nil {
		t.Fatal()
	}

	hash := l.Latest(ac.Address, token.Type)

	if hash != block.GetHash() {
		t.Fatal("err")
	}
}

func TestLedgerSession_Account(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	token := mock.TokenMeta(block.GetAddress())
	token.Type = block.GetToken()
	token2 := mock.TokenMeta(block.GetAddress())
	token2.Type = block.GetToken()
	ac := types.AccountMeta{Address: token.BelongTo, Tokens: []*types.TokenMeta{token, token2}}
	if err := l.AddAccountMeta(&ac); err != nil {
		t.Fatal()
	}

	am, err := l.Account(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(am.Tokens))
}

func TestLedgerSession_Token(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	block := addStateBlock(t, l)
	token := mock.TokenMeta(block.GetAddress())
	token.Type = block.GetToken()
	ac := types.AccountMeta{Address: token.BelongTo, Tokens: []*types.TokenMeta{token}}
	if err := l.AddAccountMeta(&ac); err != nil {
		t.Fatal()
	}

	tm, err := l.Token(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*tm)
}

func TestLedgerSession_Pending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, _ := addPending(t, l)
	pending, err := l.Pending(pendingkey.Address)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range pending {
		t.Log(k, v)
	}
}

func TestLedgerSession_Balance(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	am := addAccountMeta(t, l)
	balances, err := l.Balance(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range balances {
		t.Log(k, v)
	}
}

func TestLedgerSession_TokenBalance(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := addTokenMeta(t, l)
	balance, err := l.TokenBalance(tm.BelongTo, tm.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(balance)
}

func TestLedgerSession_TokenPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pendingkey, pendinginfo := addPending(t, l)
	pending, err := l.TokenPending(pendingkey.Address, pendinginfo.Type)
	if err != nil && err != ErrPendingNotFound {
		t.Fatal(err)
	}
	for k, v := range pending {
		t.Log(k, v)
	}
}

func TestLedger_AddOrUpdatePerformance(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	for i := 0; i < 20; i++ {
		pt := types.NewPerformanceTime()
		pt.Hash = mock.Hash()
		err := l.AddOrUpdatePerformance(pt)
		if err != nil {
			t.Fatal(err)
		}
	}

	err := l.PerformanceTimes(func(p *types.PerformanceTime) {
		t.Logf("%s", p.String())
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddOrUpdatePerformance2(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	pt := types.NewPerformanceTime()
	h := mock.Hash()
	pt.Hash = h

	err := l.AddOrUpdatePerformance(pt)
	if err != nil {
		t.Fatal(err)
	}

	t3 := time.Now().AddDate(0, 0, 1).Unix()
	pt.T3 = t3

	err = l.AddOrUpdatePerformance(pt)
	if err != nil {
		t.Fatal(err)
	}

	exist, err := l.IsPerformanceTimeExist(h)
	if err != nil {
		t.Fatal(err)
	}
	if !exist {
		t.Fatal("error exist")
	}

	pt2, err := l.GetPerformanceTime(h)
	if err != nil {
		t.Fatal(err)
	}

	if pt2.T3 != t3 {
		t.Fatal("err t3z")
	}

	b, err := l.IsPerformanceTimeExist(types.ZeroHash)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Fatal("error exist2")
	}

}

func TestLedger_OnlineRepresentations(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	var addr []*types.Address
	for i := 0; i < 10; i++ {
		a1 := mock.Address()
		addr = append(addr, &a1)
	}

	err := l.SetOnlineRepresentations(addr)
	if err != nil {
		t.Fatal(err)
	}

	if addr2, err := l.GetOnlineRepresentations(); err == nil {
		if len(addr) != len(addr2) || len(addr2) != 10 {
			t.Fatal("invalid online rep size")
		}
		for i, v := range addr {
			if v.String() != addr2[i].String() {
				t.Fatal("invalid ")
			} else {
				t.Log(v.String())
			}
		}
	} else {
		t.Fatal(err)
	}
}

func TestLedger_SetOnlineRepresentations(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	var addr []*types.Address
	err := l.SetOnlineRepresentations(addr)
	if err != nil {
		t.Fatal(err)
	}
	if addr2, err := l.GetOnlineRepresentations(); err == nil {
		if len(addr2) != 0 {
			t.Fatal("invalid online rep")
		}
	} else {
		t.Fatal(err)
	}
}

func TestLedger_BlockChild(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addr1 := mock.Address()
	addr2 := mock.Address()
	b1 := mock.StateBlockWithoutWork()
	b1.Address = addr1

	b2 := mock.StateBlockWithoutWork()
	b2.Address = addr1
	b2.Type = types.Send
	b2.Previous = b1.GetHash()

	b3 := mock.StateBlockWithoutWork()
	b3.Address = addr2
	b3.Link = b1.GetHash()

	b4 := mock.StateBlockWithoutWork()
	b4.Address = addr1
	b4.Type = types.Send
	b4.Previous = b1.GetHash()

	if err := l.AddStateBlock(b1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(b2); err != nil {
		t.Fatal(err)
	}
	h, err := l.GetChild(b1.GetHash(), b2.GetAddress())
	if err != nil {
		t.Fatal(err)
	}
	if h != b2.GetHash() {
		t.Fatal()
	}

	if err := l.AddStateBlock(b3); err != nil {
		t.Fatal(err)
	}
	h, err = l.GetChild(b1.GetHash(), b3.GetAddress())
	if err != nil {
		t.Fatal(err)
	}
	if h != b3.GetHash() {
		t.Fatal()
	}

	if err := l.AddStateBlock(b4); err == nil {
		t.Fatal()
	}

	if err := l.DeleteStateBlock(b2.GetHash()); err != nil {
		t.Fatal(err)
	}

	h, err = l.GetChild(b1.GetHash(), b2.GetAddress())
	if err != nil {
		t.Log(err)
	}

	if err := l.AddStateBlock(b4); err != nil {
		t.Fatal(err)
	}

	h, err = l.GetChild(b1.GetHash(), b4.GetAddress())
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_MessageInfo(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	h := mock.Hash()
	m := []byte{1, 2, 3}
	if err := l.AddMessageInfo(h, m); err != nil {
		t.Fatal(err)
	}
	m2, err := l.GetMessageInfo(h)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m, m2) {
		t.Fatal("wrong result")
	}
}
