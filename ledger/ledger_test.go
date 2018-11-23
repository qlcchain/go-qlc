package ledger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/json-iterator/go"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestLedger_Empty(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	err = l.View(func() error {
		r, err := l.Empty()
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
func TestLedger_AddBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	blks := parseBlocks(t, "testdata/blocks.json")
	err = l.Update(func() error {
		for _, b := range blks {
			if err = l.AddBlock(b); err != nil {
				t.Fatal(err)
			}
		}
		if err = l.AddBlock(generateBlock()); err != nil {
			return err
		}
		if err = l.AddBlock(generateBlock()); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	h := types.Hash{}
	h.Of("f42eadd5e3316197bb1757feb9714c21a7f87cc3d32a415bd3cebbb4499ea9b2")
	err = l.View(func() error {
		blk, err := l.GetBlock(h)
		t.Log("blk,", blk)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetBlocks(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	err = l.View(func() error {
		blks, err := l.GetBlocks()
		for index, b := range blks {
			t.Log(index, b)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	h := types.Hash{}
	h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")
	err = l.Update(func() error {
		return l.DeleteBlock(h)
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	h := types.Hash{}
	h.Of("f464d89184c7a9046cadabc4b8bc40782e0147b61f30f8a8b01e533b0566df1c")
	err = l.View(func() error {
		r, err := l.HasBlock(h)
		t.Log("hasblock,", r)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_CountBlocks(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	err = l.View(func() error {
		r, err := l.CountBlocks()
		t.Log("blk count,", r)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRandomBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	err = l.View(func() error {
		b, err := l.GetRandomBlock()
		t.Log("blk ,", b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
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
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, blk, kind := parseUncheckedBlock(t)
	err = l.Update(func() error {
		err = l.AddUncheckedBlock(parentHash, blk, kind)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetUncheckedBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, _, kind := parseUncheckedBlock(t)
	err = l.Update(func() error {
		b, err := l.GetUncheckedBlock(parentHash, kind)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("unchecked,", b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_CountUncheckedBlocks(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	err = l.Update(func() error {
		c, err := l.CountUncheckedBlocks()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("unchecked count,", c)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasUncheckedBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, _, kind := parseUncheckedBlock(t)

	err = l.Update(func() error {
		r, err := l.HasUncheckedBlock(parentHash, kind)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("has unchecked,", r)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteUncheckedBlock(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	parentHash, _, kind := parseUncheckedBlock(t)

	err = l.Update(func() error {
		err := l.DeleteUncheckedBlock(parentHash, kind)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
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
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	accountMetas := parseAccountMetas(t, "testdata/account.json")
	err = l.Update(func() error {
		for _, a := range accountMetas {
			fmt.Println(a)
			err := l.AddAccountMeta(a)
			if err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetAccountMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	err = l.View(func() error {
		a, err := l.GetAccountMeta(address)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("account,", a)
		for _, token := range a.Tokens {
			t.Log("token,", token)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasAccountMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	err = l.View(func() error {
		r, err := l.HasAccountMeta(address)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("has account,", r)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteAccountMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")

	err = l.Update(func() error {
		err := l.DeleteAccountMeta(address)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddOrUpdateAccountMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	accountMetas := parseAccountMetas(t, "testdata/accountupdate.json")
	err = l.Update(func() error {
		for _, a := range accountMetas {
			err := l.AddOrUpdateAccountMeta(a)
			if err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_UpdateAccountMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	accountMetas := parseAccountMetas(t, "testdata/accountupdate.json")
	err = l.Update(func() error {
		for _, a := range accountMetas {
			err := l.UpdateAccountMeta(a)
			if err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})
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
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	tokenmeta, address, _ := parseToken(t)

	err = l.Update(func() error {
		err = l.AddTokenMeta(address, &tokenmeta)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetTokenMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	_, address, tokenType := parseToken(t)

	err = l.View(func() error {
		token, err := l.GetTokenMeta(address, tokenType)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("token,", token)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_AddOrUpdateTokenMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	tokenmeta, address, _ := parseToken(t)

	err = l.Update(func() error {
		err = l.AddOrUpdateTokenMeta(address, &tokenmeta)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_UpdateTokenMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	tokenmeta, address, _ := parseToken(t)

	err = l.Update(func() error {
		err = l.UpdateTokenMeta(address, &tokenmeta)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DelTokenMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	_, address, tokentype := parseToken(t)

	err = l.Update(func() error {
		err = l.DeleteTokenMeta(address, tokentype)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_HasTokenMeta(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	_, address, tokenType := parseToken(t)

	err = l.View(func() error {
		r, err := l.HasTokenMeta(address, tokenType)
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

func TestLedger_AddRepresentationWeight(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount, err := types.ParseBalance("400.004", "Mqlc")
	err = l.Update(func() error {
		err := l.AddRepresentationWeight(address, amount)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_SubRepresentationWeight(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	amount, err := types.ParseBalance("100.004", "Mqlc")
	err = l.Update(func() error {
		err := l.SubRepresentationWeight(address, amount)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetRepresentation(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	err = l.Update(func() error {
		a, err := l.GetRepresentation(address)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("amount,", a)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
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
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, hash, pendinfo := parsePending(t)

	err = l.Update(func() error {
		err = l.AddPending(address, hash, &pendinfo)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetPending(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()

	address, hash, _ := parsePending(t)
	err = l.View(func() error {
		p, err := l.GetPending(address, hash)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("pending,", p)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeletePending(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	address, hash, _ := parsePending(t)

	err = l.Update(func() error {
		err = l.DeletePending(address, hash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func parseFrontier(t *testing.T) (frontier types.Frontier) {
	hash := "391cf191094c40f0b68e2e5f75f6bee92a2e0bd93ceaa4a6738db9f19b728948"
	address := "qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby"
	frontier.Hash.Of(hash)
	if address, err := types.HexToAddress(address); err != nil {
		t.Fatal(err)
	} else {
		frontier.Address = address
	}
	return
}
func TestLedger_AddFrontier(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	frontier := parseFrontier(t)
	err = l.Update(func() error {
		err = l.AddFrontier(&frontier)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetFrontier(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	frontier := parseFrontier(t)
	err = l.View(func() error {
		f, err := l.GetFrontier(frontier.Hash)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("frontier,", f)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_GetFrontiers(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	err = l.View(func() error {
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
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestLedger_DeleteFrontier(t *testing.T) {
	l, err := NewLedger()
	if err != nil {
		t.Fatal(err)

	}
	defer l.Close()
	frontier := parseFrontier(t)

	err = l.Update(func() error {
		err = l.DeleteFrontier(frontier.Hash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
