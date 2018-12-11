package ledger

import (
	"encoding/json"
	"fmt"
	"github.com/qlcchain/go-qlc/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
)

var (
	l *Ledger
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger")
	l = NewLedger(dir)

	return func(t *testing.T) {
		err := l.db.Erase()
		//err := l.Close()
		if err != nil {
			t.Fatal(err)
		}
		//err = os.RemoveAll(dir)
		//if err != nil {
		//	t.Fatal(err)
		//}
	}
}

func TestLedger_Instance1(t *testing.T) {
	dir := filepath.Join(config.QlcTestDataDir(), "ledger1")
	l1 := NewLedger(dir)
	l2 := NewLedger(dir)
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
	if l1 == nil || l2 == nil || reflect.DeepEqual(l1, l2) {
		t.Fatal("error")
	}
	_ = os.RemoveAll(dir)
	_ = os.RemoveAll(dir2)
}

func TestLedger_Empty(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	e, err := session.Empty()

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("is empty %s", strconv.FormatBool(e))
}

func TestLedger_NewLedgerSession1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	s1 := l.NewLedgerSession(false)
	defer s1.Close()

	s2 := l.NewLedgerSession(false)
	defer s2.Close()

	if s1 == s2 {
		t.Fatal("s1==s2")
	}

	if s1.getTxn(false) == s2.getTxn(false) {
		t.Fatal("s1_txn==s2_txn")
	}
}

func TestLedger_NewLedgerSession2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	s1 := l.NewLedgerSession(true)
	defer s1.Close()

	s2 := l.NewLedgerSession(true)
	defer s2.Close()

	if s1 == s2 {
		t.Fatal("s1==s2")
	}

	if s1.getTxn(false) == s2.getTxn(false) {
		t.Fatal("s1_txn==s2_txn")
	}

	if s1.getTxn(true) != s1.getTxn(true) {
		t.Fatal("s1_txn1!=s1_txn2")
	}
	if s1.getTxn(false) != s1.getTxn(false) {
		t.Fatal("s1_txn1!=s1_txn2")
	}

	if s1.getTxn(true) == s1.getTxn(false) {
		t.Fatal("s1_txn1==s1_txn2'")
	}
}

func TestLedgerSession_BatchUpdate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	s1 := l.NewLedgerSession(true)
	defer s1.Close()

	ams := parseAccountMetas(t, "testdata/account.json")
	tm, address, tokenType := parseToken(t)

	err := s1.BatchUpdate(func() error {
		for _, a := range ams {
			err := s1.AddAccountMeta(a)
			if err != nil {
				return err
			}
		}
		err := s1.AddTokenMeta(address, &tm)
		if err != nil {
			return err
		}
		r, err := s1.HasTokenMeta(address, tokenType)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("has token,", r)
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}
}

func generateBlock() types.Block {
	var blk types.StateBlock
	random.Bytes(blk.PreviousHash[:])
	random.Bytes(blk.Representative[:])
	random.Bytes(blk.Address[:])
	random.Bytes(blk.Signature[:])
	random.Bytes(blk.Link[:])
	random.Bytes(blk.Signature[:])
	random.Bytes(blk.Token[:])
	return &blk
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
			var json = jsoniter.ConfigCompatibleWithStandardLibrary
			var blk types.StateBlock
			if err := json.Unmarshal(data, &blk); err != nil {
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
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	s := l.NewLedgerSession(false)
	defer s.Close()

	blks := parseBlocks(t, "testdata/blocks.json")

	if err := s.AddBlock(blks[0]); err != nil {
		t.Log(err)
	}
}

func TestLedger_GetBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	blks := parseBlocks(t, "testdata/blocks.json")
	if err := session.AddBlock(blks[0]); err != nil {
		t.Fatal(err)
	}

	h := blks[0].GetHash()

	blk, err := session.GetBlock(h)
	t.Log("blk,", blk)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetAllBlocks(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	r, err := session.CountBlocks()
	t.Log("blk count, ", r)

	blks, err := session.GetBlocks()
	for index, b := range blks {
		t.Log(index, b)
	}

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	h := types.Hash{}
	h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")

	session := l.NewLedgerSession(false)
	defer session.Close()

	err := session.DeleteBlock(h)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	h := types.Hash{}
	h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")

	session := l.NewLedgerSession(false)
	defer session.Close()

	r, err := session.HasBlock(h)
	t.Log("hasblock,", r)

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRandomBlock_Empty(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	b, err := session.GetRandomBlock()

	if err != ErrStoreEmpty {
		t.Fatal(err)
	}
	t.Log("blk ,", b)
}

func parseUncheckedBlock(t *testing.T) (parentHash types.Hash, blk types.Block, kind types.UncheckedKind) {
	parentHash.Of("d66750ccbb0ff65db134efaaec31d0b123a557df34e7e804d6884447ee589b3c")
	blk, _ = types.NewBlock(byte(types.State))
	blocks := parseBlocks(t, "testdata/uncheckedblock.json")
	fmt.Println(blocks)
	blk = blocks[0]
	kind = types.UncheckedKindLink
	return
}

func addUncheckedBlock(t *testing.T, ls *LedgerSession) {
	parentHash, blk, kind := parseUncheckedBlock(t)
	if err := ls.AddUncheckedBlock(parentHash, blk, kind); err != nil && err != ErrUncheckedBlockExists {
		t.Fatal(err)
	}
}
func TestLedger_AddUncheckedBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addUncheckedBlock(t, session)
}
func TestLedger_GetUncheckedBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	parentHash, _, kind := parseUncheckedBlock(t)
	addUncheckedBlock(t, session)

	if b, err := session.GetUncheckedBlock(parentHash, kind); err != nil && err != ErrUncheckedBlockNotFound {
		t.Fatal(err)
	} else {
		t.Log("unchecked,", b)
	}

}
func TestLedger_CountUncheckedBlocks(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	c, err := session.CountUncheckedBlocks()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("unchecked count,", c)
}
func TestLedger_HasUncheckedBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	parentHash, _, kind := parseUncheckedBlock(t)

	r, err := session.HasUncheckedBlock(parentHash, kind)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has unchecked,", r)
}
func TestLedger_DeleteUncheckedBlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	parentHash, _, kind := parseUncheckedBlock(t)
	err := session.DeleteUncheckedBlock(parentHash, kind)
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
		var accountmeta types.AccountMeta
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err := json.Unmarshal(data, &accountmeta); err != nil {
			t.Fatal(err)
		}
		accountmetas = append(accountmetas, &accountmeta)
	}

	return
}

func addAccountMeta(t *testing.T, ls *LedgerSession) {
	ams := parseAccountMetas(t, "testdata/account.json")
	for _, a := range ams {
		err := ls.AddAccountMeta(a)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_AddAccountMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addAccountMeta(t, session)
}
func TestLedger_GetAccountMeta_Empty(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	_, err := session.GetAccountMeta(address)
	if err != ErrAccountNotFound {
		t.Fatal(err)
	}
}
func TestLedger_GetAccountMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addAccountMeta(t, session)
	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	a, err := session.GetAccountMeta(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", a)
	for _, token := range a.Tokens {
		t.Log("token,", token)
	}
}
func TestLedger_HasAccountMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	r, err := session.HasAccountMeta(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has account,", r)
}
func TestLedger_DeleteAccountMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")

	err := session.DeleteAccountMeta(address)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddOrUpdateAccountMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	ams := parseAccountMetas(t, "testdata/accountupdate.json")
	for _, a := range ams {
		err := session.AddOrUpdateAccountMeta(a)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_UpdateAccountMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addAccountMeta(t, session)
	ams := parseAccountMetas(t, "testdata/accountupdate.json")
	for _, a := range ams {
		err := session.UpdateAccountMeta(a)
		if err != nil {
			t.Fatal(err)
		}
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
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err := json.Unmarshal(data, &tokenmeta); err != nil {
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

func addTokenMeta(t *testing.T, ls *LedgerSession) {
	addAccountMeta(t, ls)

	tm, address, _ := parseToken(t)

	err := ls.AddTokenMeta(address, &tm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddTokenMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addTokenMeta(t, session)
}
func TestLedger_GetTokenMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addTokenMeta(t, session)

	_, address, tokenType := parseToken(t)

	token, err := session.GetTokenMeta(address, tokenType)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token,", token)
}
func TestLedger_AddOrUpdateTokenMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addTokenMeta(t, session)

	tm, address, _ := parseToken(t)

	err := session.AddOrUpdateTokenMeta(address, &tm)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_UpdateTokenMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addTokenMeta(t, session)
	tm, address, _ := parseToken(t)

	err := session.UpdateTokenMeta(address, &tm)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DelTokenMeta(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addTokenMeta(t, session)
	_, address, tokentype := parseToken(t)

	err := session.DeleteTokenMeta(address, tokentype)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasTokenMeta_False(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	address, _, _ := types.GenerateAddress()
	tokenType := types.Hash{}
	addTokenMeta(t, session)

	if _, err := session.HasTokenMeta(address, tokenType); err == ErrAccountNotFound {

	} else {
		t.Fatal(err)
	}
}

func TestLedger_HasTokenMeta_True(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	_, address, tokenType := parseToken(t)

	addTokenMeta(t, session)
	r, err := session.HasTokenMeta(address, tokenType)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", r)
}

func addRepresentationWeight(t *testing.T, s *LedgerSession) {
	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount, err := types.ParseBalance("400.004", "Mqlc")

	err = s.AddRepresentation(address, amount)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddRepresentationWeight(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()

	addRepresentationWeight(t, session)
}
func TestLedger_SubRepresentationWeight(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addRepresentationWeight(t, session)

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount, err := types.ParseBalance("100.004", "Mqlc")
	err = session.SubRepresentation(address, amount)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRepresentation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addRepresentationWeight(t, session)

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	a, err := session.GetRepresentation(address)
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
	hash.Of("a624942c313e8ddd7bc12cf6188e4fb9d10da4238086aceca7f81ea3fc595ba9")

	balance, err := types.ParseBalance("2345.6789", "Mqlc")
	if err != nil {
		t.Fatal(err)
	}
	typehash := types.Hash{}
	typehash.Of("191cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b722448")
	pendinginfo = types.PendingInfo{
		Source: address,
		Amount: balance,
		Type:   typehash,
	}
	return
}

func addPending(t *testing.T, s *LedgerSession) {
	address, hash, pendinfo := parsePending(t)

	err := s.AddPending(address, hash, &pendinfo)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddPending(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addPending(t, session)
}
func TestLedger_GetPending(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	addPending(t, session)
	address, hash, _ := parsePending(t)
	p, err := session.GetPending(address, hash)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pending,", p)
}
func TestLedger_DeletePending(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	address, hash, _ := parsePending(t)

	err := session.DeletePending(address, hash)
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
	random.Bytes(frontier.HeaderBlock[:])
	random.Bytes(frontier.OpenBlock[:])
	return &frontier
}
func TestLedger_AddFrontier(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	frontier := parseFrontier(t)
	if err := session.AddFrontier(&frontier); err != nil && err != ErrFrontierExists {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		err := session.AddFrontier(generateFrontier())
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_GetFrontier(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	frontier := parseFrontier(t)
	f, err := session.GetFrontier(frontier.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier,", f)
}
func TestLedger_GetAllFrontiers(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	c, err := session.CountFrontiers()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier count,", c)

	fs, err := session.GetFrontiers()
	if err != nil {
		t.Fatal(err)
	}
	for index, f := range fs {
		t.Log("frontier", index, f)
	}
}
func TestLedger_DeleteFrontier(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	session := l.NewLedgerSession(false)
	defer session.Close()
	frontier := parseFrontier(t)

	err := session.DeleteFrontier(frontier.HeaderBlock)
	if err != nil {
		t.Fatal(err)
	}
}
