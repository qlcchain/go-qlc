package ledger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
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
		//err := l.db.Erase()
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
	_ = os.RemoveAll(dir)
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
	txn := l.db.NewTransaction(false)
	txn2, flag := l.getTxn(false, txn)
	if flag {
		t.Fatal("get txn flag error")
		if txn != txn2 {
			t.Fatal("txn!=tnx2")
		}
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
		addAccountMeta(t, l)
		am := getAccount(t)
		a, err := l.GetAccountMeta(am.Address)
		if err != nil {
			t.Fatal(err)
		}
		logger.Info(a)
		return nil

	})

	if err != nil {
		t.Fatal(err)
	}
}

func parseBlocks(t *testing.T, filename string) (blocks []types.Block) {
	type fileStruct struct {
		Blocks []json.RawMessage `json:"blocks"`
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	var file fileStruct
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}

	for _, data := range file.Blocks {
		var values map[string]interface{}
		if err = json.Unmarshal(data, &values); err != nil {
			t.Fatal(err)
		}

		id, ok := values["type"]
		if !ok {
			t.Fatalf("no 'type' key found in block")
		}
		//var blk types.Block
		switch id {
		case "state":
			var blk types.StateBlock
			if err := jsoniter.Unmarshal(data, &blk); err != nil {
				t.Fatal(err)
			} else {
				blocks = append(blocks, &blk)
			}
		case types.SmartContract:
			//blk := new(types.SmartContractBlock)
		default:
			t.Fatalf("unsupported block type")
		}
	}
	return
}
func TestLedger_AddBlockWithSingleTxn(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	blks := parseBlocks(t, "testdata/blocks.json")

	if err := l.AddBlock(blks[0]); err != nil {
		t.Log(err)
	}
}

func addBlocks(t *testing.T, l *Ledger) {
	blks := parseBlocks(t, "testdata/blocks.json")
	if err := l.AddBlock(blks[0]); err != nil && err != ErrBlockExists {
		t.Fatal(err)
	}
}

func TestLedger_GetBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addBlocks(t, l)

	blks := parseBlocks(t, "testdata/blocks.json")
	h := blks[0].GetHash()

	blk, err := l.GetBlock(h)
	t.Log("blk,", blk)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetAllBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addBlocks(t, l)
	r, err := l.CountBlocks()
	t.Log("blk count, ", r)

	blks, err := l.GetBlocks()
	for index, b := range blks {
		t.Log(index, b, *b)
	}

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	h := types.Hash{}
	_ = h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")

	addBlocks(t, l)
	err := l.DeleteBlock(h)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	h := types.Hash{}
	_ = h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")

	r, err := l.HasBlock(h)
	t.Log("hasblock,", r)

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRandomBlock_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	b, err := l.GetRandomBlock()

	if err != ErrStoreEmpty {
		t.Fatal(err)
	}
	t.Log("blk ,", b)
}

func parseUncheckedBlock(t *testing.T) (parentHash types.Hash, blk types.Block, kind types.UncheckedKind) {
	_ = parentHash.Of("d66750ccbb0ff65db134efaaec31d0b123a557df34e7e804d6884447ee589b3c")
	blk, _ = types.NewBlock(types.State)
	blocks := parseBlocks(t, "testdata/uncheckedblock.json")
	fmt.Println(blocks)
	blk = blocks[0]
	kind = types.UncheckedKindLink
	return
}

func addUncheckedBlock(t *testing.T, l *Ledger) {
	parentHash, blk, kind := parseUncheckedBlock(t)
	if err := l.AddUncheckedBlock(parentHash, blk, kind); err != nil && err != ErrUncheckedBlockExists {
		t.Fatal(err)
	}
}
func TestLedger_AddUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addUncheckedBlock(t, l)
}
func TestLedger_GetUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := parseUncheckedBlock(t)
	addUncheckedBlock(t, l)

	if b, err := l.GetUncheckedBlock(parentHash, kind); err != nil && err != ErrUncheckedBlockNotFound {
		t.Fatal(err)
	} else {
		t.Log("unchecked,", b)
	}

}
func TestLedger_CountUncheckedBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	c, err := l.CountUncheckedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("unchecked count,", c)
}
func TestLedger_HasUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := parseUncheckedBlock(t)

	r, err := l.HasUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has unchecked,", r)
}
func TestLedger_DeleteUncheckedBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	parentHash, _, kind := parseUncheckedBlock(t)
	err := l.DeleteUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
}

func parseAccountMetas(t *testing.T, filename string) (accountmetas []*types.AccountMeta) {
	type fileStruct struct {
		AccountMetas []json.RawMessage `json:"accountmetas"`
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var file fileStruct
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	for _, data := range file.AccountMetas {
		var am types.AccountMeta
		if err := jsoniter.Unmarshal(data, &am); err != nil {
			t.Fatal(err)
		}
		accountmetas = append(accountmetas, &am)
	}

	return
}

func getAccount(t *testing.T) *types.AccountMeta {
	const seed = "5a32b2325437cc10c07e36161fcda24f01ec0038969ecaaa709a133000bf4b94"
	_, priv, err := types.KeypairFromSeed(seed, 1)
	if err != nil {
		t.Fatal()
	}
	ac1 := types.NewAccount(priv)
	am := mock.AccountMeta(ac1.Address())
	return am
}

func addAccountMeta(t *testing.T, l *Ledger) {
	am := getAccount(t)
	if err := l.AddAccountMeta(am); err != nil {
		t.Fatal()
	}
}

func TestLedger_AddAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addAccountMeta(t, l)
}

func TestLedger_GetAccountMeta_Empty(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	_, err := l.GetAccountMeta(address)
	if err != ErrAccountNotFound {
		t.Fatal(err)
	}
}
func TestLedger_GetAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	am := getAccount(t)
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

	addAccountMeta(t, l)
	am := getAccount(t)
	r, err := l.HasAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has account,", r)
}

func TestLedger_DeleteAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	am := getAccount(t)
	err := l.DeleteAccountMeta(am.Address)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddOrUpdateAccountMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addAccountMeta(t, l)
	am := getAccount(t)
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
	addAccountMeta(t, l)
	am := getAccount(t)
	token := mock.TokenMeta(am.Address)
	am.Tokens = append(am.Tokens, token)

	err := l.AddOrUpdateAccountMeta(am)
	if err != nil {
		t.Fatal(err)
	}
}

func parseToken(t *testing.T) (tokenmeta types.TokenMeta, address types.Address, tokenType types.Hash) {
	filename := "testdata/token.json"
	type fileStruct struct {
		TokenMetas []json.RawMessage `json:"tokens"`
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var file fileStruct
	if err = json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}

	//only get one token
	for _, data := range file.TokenMetas {
		if err := jsoniter.Unmarshal(data, &tokenmeta); err != nil {
			t.Fatal(err)
		}
		break
	}

	address, err = types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	tokenType = tokenmeta.Type
	return

}

func getToken(addr types.Address) *types.TokenMeta {
	token := mock.TokenMeta(addr)
	return token

}

func addTokenMeta(t *testing.T, l *Ledger, token *types.TokenMeta) {
	addAccountMeta(t, l)
	if err := l.AddTokenMeta(token.BelongTo, token); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)
	addTokenMeta(t, l, token)
}

func TestLedger_GetTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)
	addTokenMeta(t, l, token)

	token, err := l.GetTokenMeta(tm.Address, token.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token,", token)
}
func TestLedger_AddOrUpdateTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)
	addTokenMeta(t, l, token)

	token2 := mock.TokenMeta(tm.Address)

	err := l.AddOrUpdateTokenMeta(tm.Address, token2)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_UpdateTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)
	addTokenMeta(t, l, token)

	token2 := mock.TokenMeta(tm.Address)
	token2.Type = token.Type

	err := l.UpdateTokenMeta(tm.Address, token2)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DelTokenMeta(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)
	addTokenMeta(t, l, token)

	err := l.DeleteTokenMeta(tm.Address, token.Type)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasTokenMeta_False(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address, _, _ := types.GenerateAddress()
	tokenType := types.Hash{}
	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)

	addTokenMeta(t, l, token)

	has, err := l.HasTokenMeta(address, tokenType)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", has)
}

func TestLedger_HasTokenMeta_True(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	tm := getAccount(t)
	token := mock.TokenMeta(tm.Address)
	addTokenMeta(t, l, token)

	r, err := l.HasTokenMeta(tm.Address, token.Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", r)
}

func addRepresentationWeight(t *testing.T, l *Ledger) {
	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount := types.StringToBalance("400004")

	err := l.AddRepresentation(address, amount)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddRepresentationWeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)
	addRepresentationWeight(t, l)
}

func TestLedger_SubRepresentationWeight(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addRepresentationWeight(t, l)

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount := types.StringToBalance("100004")
	err := l.SubRepresentation(address, amount)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRepresentation(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addRepresentationWeight(t, l)

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	a, err := l.GetRepresentation(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("amount,", a)
}

func parsePending(t *testing.T) (address types.Address, hash types.Hash, pendinginfo types.PendingInfo) {
	address, err := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	_ = hash.Of("a624942c313e8ddd7bc12cf6188e4fb9d10da4238086aceca7f81ea3fc595ba9")

	balance := types.StringToBalance("23456789")
	typehash := types.Hash{}
	_ = typehash.Of("191cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b722448")
	pendinginfo = types.PendingInfo{
		Source: address,
		Amount: balance,
		Type:   typehash,
	}
	return
}

func addPending(t *testing.T, l *Ledger) {
	address, hash, pendinfo := parsePending(t)

	err := l.AddPending(types.PendingKey{Address: address, Hash: hash}, &pendinfo)
	if err != nil {
		t.Fatal(err)
	}
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
	address, hash, _ := parsePending(t)
	p, err := l.GetPending(types.PendingKey{Address: address, Hash: hash})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pending,", p)
}
func TestLedger_DeletePending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	address, hash, _ := parsePending(t)

	err := l.DeletePending(types.PendingKey{Address: address, Hash: hash})
	if err != nil {
		t.Fatal(err)
	}
}

func parseFrontier(t *testing.T) (frontier types.Frontier) {
	headerhash := "391cf191094c40f0b68e2e5f75f6bee92a2e0bd93ceaa4a6738db9f19b728948"
	openhash := "001cf191094c40f0b68e2e5f75f6bee92a2e0bd93ceaa4a6738db9f19b728948"
	frontier.HeaderBlock.Of(headerhash)
	frontier.OpenBlock.Of(openhash)
	return
}
func generateFrontier() *types.Frontier {
	var frontier types.Frontier
	_ = random.Bytes(frontier.HeaderBlock[:])
	_ = random.Bytes(frontier.OpenBlock[:])
	return &frontier
}
func TestLedger_AddFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	frontier := parseFrontier(t)
	if err := l.AddFrontier(&frontier); err != nil && err != ErrFrontierExists {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err := l.AddFrontier(generateFrontier())
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_GetFrontier(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	frontier := parseFrontier(t)
	if err := l.AddFrontier(&frontier); err != nil && err != ErrFrontierExists {
		t.Fatal(err)
	}
	f, err := l.GetFrontier(frontier.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier,", f)
}

func TestLedger_GetAllFrontiers(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

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

	frontier := parseFrontier(t)

	err := l.DeleteFrontier(frontier.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
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

	addAccountMeta(t, l)
	block := mock.StateBlock()
	ac := mock.AccountMeta(block.GetAddress())
	ac.Tokens[0].Header = block.GetHash()
	ac.Tokens[0].Type = block.(*types.StateBlock).Token
	l.AddAccountMeta(ac)
	l.AddBlock(block)

	hash := l.Latest(ac.Address, ac.Tokens[0].Type)

	if hash != block.GetHash() {
		t.Fatal("err")
	}
}

func TestLedgerSession_Account(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	block := mock.StateBlock()
	ac := mock.AccountMeta(block.GetAddress())
	l.AddAccountMeta(ac)

	l.AddBlock(block)

	am, err := l.Account(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(am.Tokens))
}

func TestLedgerSession_Token(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	block := mock.StateBlock()
	ac := mock.AccountMeta(block.GetAddress())
	ac.Tokens[0].Type = block.(*types.StateBlock).Token
	l.AddAccountMeta(ac)

	l.AddBlock(block)
	tm, err := l.Token(block.GetHash())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(*tm)
}

func TestLedgerSession_Pending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	addBlocks(t, l)

	addr, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	pending, err := l.Pending(addr)
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

	ac := getAccount(t)
	am := mock.AccountMeta(ac.Address)
	l.AddAccountMeta(am)

	balances, err := l.Balance(ac.Address)
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

	ac := getAccount(t)
	am := mock.AccountMeta(ac.Address)
	l.AddAccountMeta(am)
	balance, err := l.TokenBalance(ac.Address, am.Tokens[0].Type)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(balance)
}

func TestLedgerSession_TokenPending(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	addAccountMeta(t, l)
	addBlocks(t, l)

	addr, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	token := types.Hash{}
	_ = token.Of("991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728918")
	pending, err := l.TokenPending(addr, token)
	if err != nil && err != ErrPendingNotFound {
		t.Fatal(err)
	}
	t.Log(pending)
}

func TestLedger_Rollback(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	bs, err := mock.BlockChain()
	if err != nil {
		t.Fatal()
	}
	for _, b := range bs {
		processBlock(t, l, b)
	}
	h := bs[2].GetHash()
	if err := l.Rollback(h); err != nil {
		t.Fatal()
	}

	//check(t, l)
}

func check(t *testing.T, l *Ledger) {
	blocks, _ := l.GetBlocks()
	fmt.Println("----blocks: ")
	for _, b := range blocks {
		fmt.Println(*b)
	}

	fmt.Println("----frontiers:")
	fs, _ := l.GetFrontiers()
	for _, f := range fs {
		fmt.Println(f)
	}
	fmt.Println("----account: ")
	var addrs []types.Address
	addr1, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	addrs = append(addrs, addr1)
	addr2, _ := types.HexToAddress("qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i")
	addrs = append(addrs, addr2)
	addr3, _ := types.HexToAddress("qlc_3pu4ggyg36nienoa9s9x95a615m1natqcqe7bcrn3t3ckq1srnnkh8q5xst5")
	addrs = append(addrs, addr3)
	for index, addr := range addrs {
		if ac, err := l.GetAccountMeta(addr); err == nil {
			fmt.Println("   account ", index, " ", ac.Address)
			for _, t := range ac.Tokens {
				fmt.Println("       token ", t)
			}
		}
	}
	fmt.Println("----representation:")
	for _, addr := range addrs {
		if b, err := l.GetRepresentation(addr); err == nil {
			fmt.Println(addr, b)
		}

	}
}
