package ledger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/ledger/db"
)

var dir_testdb = util.QlcDir("test", "ledger")

func initTestLedger() (*Ledger, error) {
	store, err := db.NewBadgerStore(dir_testdb)
	if err != nil {
		return nil, err
	}
	ledger := Ledger{store: store}
	return &ledger, nil
}
func TestLedger_RemoveDB(t *testing.T) {
	if err := os.RemoveAll(dir_testdb); err != nil {
		t.Fatal(err)
	}
}
func TestLedger_Empty(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	err = l.GetDataInTransaction(func(txn db.StoreTxn) error {
		r, err := l.Empty(txn)
		if err != nil {
			return err
		}
		t.Log("empty,", r)
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
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	blks := parseBlocks(t, "testdata/blocks.json")

	if err = l.AddBlock(blks[0], nil); err != nil {
		t.Fatal(err)
	}

}
func TestLedger_AddBlockWithBlockTxn(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	err = l.UpdateDataInTransaction(func(txn db.StoreTxn) error {
		if err = l.AddBlock(generateBlock(), txn); err != nil {
			t.Fatal(err)
		}
		if err = l.AddBlock(generateBlock(), txn); err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	h := types.Hash{}
	h.Of("f42eadd5e3316197bb1757feb9714c21a7f87cc3d32a415bd3cebbb4499ea9b2")
	err = l.GetDataInTransaction(func(txn db.StoreTxn) error {
		blk, err := l.GetBlock(h, txn)
		t.Log("blk,", blk)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetAllBlocks(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	r, err := l.CountBlocks(nil)
	t.Log("blk count, ", r)

	blks, err := l.GetBlocks(nil)
	for index, b := range blks {
		t.Log(index, b)
	}

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	h := types.Hash{}
	h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")
	err = l.DeleteBlock(h, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	h := types.Hash{}
	h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")
	r, err := l.HasBlock(h, nil)
	t.Log("hasblock,", r)

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRandomBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	b, err := l.GetRandomBlock(nil)

	if err != nil {
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
func TestLedger_AddUncheckedBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, blk, kind := parseUncheckedBlock(t)
	err = l.AddUncheckedBlock(parentHash, blk, kind, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetUncheckedBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, _, kind := parseUncheckedBlock(t)
	b, err := l.GetUncheckedBlock(parentHash, kind, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("unchecked,", b)
}
func TestLedger_CountUncheckedBlocks(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	c, err := l.CountUncheckedBlocks(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("unchecked count,", c)
}
func TestLedger_HasUncheckedBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, _, kind := parseUncheckedBlock(t)

	r, err := l.HasUncheckedBlock(parentHash, kind, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has unchecked,", r)
}
func TestLedger_DeleteUncheckedBlock(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, _, kind := parseUncheckedBlock(t)
	err = l.DeleteUncheckedBlock(parentHash, kind, nil)
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
func TestLedger_AddAccountMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	accountMetas := parseAccountMetas(t, "testdata/account.json")
	for _, a := range accountMetas {
		err := l.AddAccountMeta(a, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_GetAccountMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	a, err := l.GetAccountMeta(address, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("account,", a)
	for _, token := range a.Tokens {
		t.Log("token,", token)
	}
}
func TestLedger_HasAccountMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	r, err := l.HasAccountMeta(address, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has account,", r)
}
func TestLedger_DeleteAccountMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")

	err = l.DeleteAccountMeta(address, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddOrUpdateAccountMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	accountMetas := parseAccountMetas(t, "testdata/accountupdate.json")
	for _, a := range accountMetas {
		err := l.AddOrUpdateAccountMeta(a, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_UpdateAccountMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	accountMetas := parseAccountMetas(t, "testdata/accountupdate.json")
	for _, a := range accountMetas {
		err := l.UpdateAccountMeta(a, nil)
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
func TestLedger_AddTokenMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	tokenmeta, address, _ := parseToken(t)

	err = l.AddTokenMeta(address, &tokenmeta, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetTokenMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	_, address, tokenType := parseToken(t)

	token, err := l.GetTokenMeta(address, tokenType, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("token,", token)
}
func TestLedger_AddOrUpdateTokenMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	tokenmeta, address, _ := parseToken(t)

	err = l.AddOrUpdateTokenMeta(address, &tokenmeta, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_UpdateTokenMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	tokenmeta, address, _ := parseToken(t)

	err = l.UpdateTokenMeta(address, &tokenmeta, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DelTokenMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	_, address, tokentype := parseToken(t)

	err = l.DeleteTokenMeta(address, tokentype, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasTokenMeta(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	_, address, tokenType := parseToken(t)

	r, err := l.HasTokenMeta(address, tokenType, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("has token,", r)
}

func TestLedger_AddRepresentationWeight(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount, err := types.ParseBalance("400.004", "Mqlc")
	err = l.AddRepresentationWeight(address, amount, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_SubRepresentationWeight(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount, err := types.ParseBalance("100.004", "Mqlc")
	err = l.SubRepresentationWeight(address, amount, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRepresentation(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	a, err := l.GetRepresentation(address, nil)
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
func TestLedger_AddPending(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, hash, pendinfo := parsePending(t)

	err = l.AddPending(address, hash, &pendinfo, nil)
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetPending(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, hash, _ := parsePending(t)
	p, err := l.GetPending(address, hash, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("pending,", p)
}
func TestLedger_DeletePending(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	address, hash, _ := parsePending(t)

	err = l.DeletePending(address, hash, nil)
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
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	frontier := parseFrontier(t)
	err = l.AddFrontier(&frontier, nil)
	if err != nil && err != ErrFrontierExists {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		err = l.AddFrontier(generateFrontier(), nil)
		if err != nil {
			t.Fatal(err)
		}
	}
}
func TestLedger_GetFrontier(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	frontier := parseFrontier(t)
	f, err := l.GetFrontier(frontier.HeaderBlock, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier,", f)
}
func TestLedger_GetAllFrontiers(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	c, err := l.CountFrontiers(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("frontier count,", c)

	fs, err := l.GetFrontiers(nil)
	if err != nil {
		t.Fatal(err)
	}
	for index, f := range fs {
		t.Log("frontier", index, f)
	}
}
func TestLedger_DeleteFrontier(t *testing.T) {
	l, err := initTestLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	frontier := parseFrontier(t)

	err = l.DeleteFrontier(frontier.HeaderBlock, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLedger_RemoveDB2(t *testing.T) {
	if err := os.RemoveAll(dir_testdb); err != nil {
		t.Fatal(err)
	}
}
