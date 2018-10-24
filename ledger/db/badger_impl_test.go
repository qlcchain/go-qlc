package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
	"io/ioutil"
	"testing"
)

const (
	dir string = "../testdatabase"
)

// Test Badger

func TestBadgerStoreTxn_Empty(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		empty, err := txn.Empty()
		if err != nil {
			return err
		}
		fmt.Println(empty)
		return nil
	})
}

// Test Badger Blocks CURD
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
			blk := new(types.StateBlock)
			blk.Type = types.State
			blk.Address, err = types.HexToAddress(values["address"].(string))
			if err != nil {
				t.Fatal(err)
			}
			blk.PreviousHash.Of(values["previousHash"].(string))
			blk.Representative, err = types.HexToAddress(values["representative"].(string))
			if err != nil {
				t.Fatal(err)
			}
			blk.Balance, err = types.ParseBalance(values["balance"].(string), "Mqlc")
			if err != nil {
				t.Fatal(err)
			}

			//rand.Seed(time.Now().UnixNano())
			//i := rand.Int()
			//link := strconv.Itoa(i) + values["link"].(string)[len(strconv.Itoa(i)):]
			//blk.Link.Of(link)
			blk.Link.Of(values["link"].(string))

			blk.Signature.Of(values["signature"].(string))
			blk.Token.Of(values["token"].(string))
			blk.Work.ParseWorkHexString(values["work"].(string))
			blocks = append(blocks, blk)
		case types.SmartContract:
			//blk := new(types.SmartContractBlock)
		default:
			t.Fatalf("unsupported block type")
		}
	}
	return
}

func TestBadgerStoreTxn_AddBlocks(t *testing.T) {
	const n = 1
	for i := 0; i < n; i++ {
		TestBadgerStoreTxn_AddBlock(t)
	}
}

func TestBadgerStoreTxn_AddBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	blocks := parseBlocks(t, "../testdata/blocks.json")
	for _, blk := range blocks {
		db.Update(func(txn StoreTxn) error {
			err := txn.AddBlock(blk)
			if err != nil {
				if err == ErrBlockExists {
					fmt.Println("warn:block already exists")
				} else {
					t.Fatal(err)
				}
			}
			return nil
		})
	}
}

func TestBadgerStoreTxn_GetBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		hash := types.Hash{}
		hash.Of("bbb23eab1706acaf717be7567c6b4568c801bfbae397503e885f8caf20e968a0")
		block, err := txn.GetBlock(hash)

		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("warn:block not found")
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(block)
			if block.Hash() != hash {
				t.Fatal(errors.New("get incorrect block"))
			}
		}
		return nil
	})
}

func TestBadgerStoreTxn_DeleteBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.Update(func(txn StoreTxn) error {
		hash := types.Hash{}
		hash.Of("bbb23eab1706acaf717be7567c6b4568c801bfbae397503e885f8caf20e968a0")
		err := txn.DeleteBlock(hash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

func TestBadgerStoreTxn_HasBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		hash := types.Hash{}
		hash.Of("f9b38dad8588db575bd81bca7a806cd9e103994b74d79931724198d99b239f8c")
		b, err := txn.HasBlock(hash)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(b)
		return nil
	})
}

func TestBadgerStoreTxn_CountBlocks(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		b, err := txn.CountBlocks()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(b)
		return nil
	})
}

func TestBadgerStoreTxn_GetRandomBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		block, err := txn.GetRandomBlock()
		if err != nil {
			if err == ErrStoreEmpty {
				fmt.Println("the store is empty")
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(block)
		}
		return nil
	})
}

// Test Badger Account CURD

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
		type accountStruct struct {
			Address string            `json:"address"`
			Tokens  []json.RawMessage `json:"tokens"`
		}

		var a accountStruct
		if err = json.Unmarshal(data, &a); err != nil {
			t.Fatal(err)
		}
		accountmeta := new(types.AccountMeta)
		accountmeta.Address, err = types.HexToAddress(a.Address)
		for _, data := range a.Tokens {
			var values map[string]interface{}
			if err = json.Unmarshal(data, &values); err != nil {
				t.Fatal(err)
			}
			token := new(types.TokenMeta)
			token.Type.Of(values["type"].(string))
			token.Header.Of(values["header"].(string))
			token.OpenBlock.Of(values["openBlock"].(string))
			token.RepBlock.Of(values["repBlock"].(string))
			token.Balance, err = types.ParseBalance(values["balance"].(string), "Mqlc")
			if err != nil {
				t.Fatal(err)
			}
			accountmeta.Tokens = append(accountmeta.Tokens, token)
		}
		accountmetas = append(accountmetas, accountmeta)
	}
	return
}

func TestBadgerStoreTxn_AddAccountMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	accountMetas := parseAccountMetas(t, "../testdata/account.json")
	for _, accountmeta := range accountMetas {
		db.Update(func(txn StoreTxn) error {
			err := txn.AddAccountMeta(accountmeta)
			if err != nil {
				if err == ErrAccountExists {
					fmt.Println("warn:account already exists")
				} else {
					t.Fatal(err)
				}
			}
			return nil
		})
	}
}

func TestBadgerStoreTxn_GetAccountMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		address, err := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
		accountmeta, err := txn.GetAccountMeta(address)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("warn:account not found")
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(accountmeta.Address)
			for _, token := range accountmeta.Tokens {
				fmt.Println(token)
			}
		}
		return nil
	})
}

func TestBadgerStoreTxn_UpdateAccountMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	accountMetas := parseAccountMetas(t, "../testdata/accountupdate.json")
	for _, accountmeta := range accountMetas {
		db.Update(func(txn StoreTxn) error {
			err := txn.UpdateAccountMeta(accountmeta)
			if err != nil {
				t.Fatal(err)
			}
			return nil
		})
	}
}

func TestBadgerStoreTxn_DeleteAccountMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.Update(func(txn StoreTxn) error {
		address, err := types.HexToAddress("qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i")
		if err != nil {
			t.Fatal(err)
		}
		err = txn.DeleteAccountMeta(address)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

// Test Badger Token CURD

func parseToken(t *testing.T) (tokenmeta types.TokenMeta, address types.Address, tokenType types.Hash) {
	filename := "../testdata/token.json"
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
	for _, data := range file.TokenMetas {
		var values map[string]interface{}
		if err = json.Unmarshal(data, &values); err != nil {
			t.Fatal(err)
		}
		tokenmeta.Type.Of(values["type"].(string))
		tokenmeta.Header.Of(values["header"].(string))
		tokenmeta.OpenBlock.Of(values["openBlock"].(string))
		tokenmeta.RepBlock.Of(values["repBlock"].(string))
		tokenmeta.Balance, err = types.ParseBalance(values["balance"].(string), "Mqlc")
		if err != nil {
			t.Fatal(err)
		}
	}

	address, err = types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	tokenType = tokenmeta.Type
	return

}

func TestBadgerStoreTxn_AddTokenMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tokenmeta, address, _ := parseToken(t)

	err = db.Update(func(txn StoreTxn) error {
		err = txn.AddTokenMeta(address, &tokenmeta)
		if err != nil {
			if err == ErrTokenExists {
				fmt.Println("warn:token already exists")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBadgerStoreTxn_GetTokenMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, address, tokenType := parseToken(t)
	db.View(func(txn StoreTxn) error {
		tokenmeta, err := txn.GetTokenMeta(address, tokenType)
		if err != nil {
			if err == ErrTokenNotFound {
				fmt.Println("warn:token not found")
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(tokenmeta)
		}
		return nil
	})
}

func TestBadgerStoreTxn_DelTokenMeta(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tokenmeta, address, _ := parseToken(t)
	err = db.Update(func(txn StoreTxn) error {
		err = txn.DelTokenMeta(address, &tokenmeta)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Test Badger Pending CURD

func parsePending(t *testing.T) (address types.Address, hash types.Hash, pendinginfo types.PendingInfo) {
	address, err := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	hash.Of("671cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b722478")

	balance, err := types.ParseBalance("2345.6789", "Mqlc")
	if err != nil {
		t.Fatal(err)
	}
	typehash := types.Hash{}
	typehash.Of("191cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b722448")
	pendinginfo = types.PendingInfo{
		address,
		balance,
		typehash,
	}
	return
}

func TestBadgerStoreTxn_AddPending(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	address, hash, pendinfo := parsePending(t)

	db.Update(func(txn StoreTxn) error {
		err := txn.AddPending(address, hash, &pendinfo)
		fmt.Println(err)
		if err != nil {
			if err == ErrPendingExists {
				fmt.Println("warn:pending transation already exists")
			} else {
				t.Fatal(err)
			}
		}
		return nil
	})
}

func TestBadgerStoreTxn_GetPending(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	address, hash, _ := parsePending(t)

	db.View(func(txn StoreTxn) error {
		pendinginfo, err := txn.GetPending(address, hash)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("warn:pending not found")
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(pendinginfo)
		}
		return nil
	})
}

func TestBadgerStoreTxn_DeletePending(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	address, hash, _ := parsePending(t)

	db.Update(func(txn StoreTxn) error {
		err = txn.DeletePending(address, hash)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

// Test Badger UncheckedBlock CURD

func parseUncheckedBlock(t *testing.T) (parentHash types.Hash, blk types.Block, kind types.UncheckedKind) {
	parentHash.Of("671cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b722471")
	blocks := parseBlocks(t, "../testdata/uncheckedblock.json")
	blk = blocks[0]
	kind = types.UncheckedKindPrevious
	return
}

func TestBadgerStoreTxn_AddUncheckedBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	parentHash, blk, kind := parseUncheckedBlock(t)
	db.Update(func(txn StoreTxn) error {
		err := txn.AddUncheckedBlock(parentHash, blk, kind)
		if err != nil {
			if err == ErrBlockExists {
				fmt.Println("warn:block already exists")
			} else {
				t.Fatal(err)
			}
		}
		return nil
	})
}

func TestBadgerStoreTxn_GetUncheckedBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	parentHash, _, kind := parseUncheckedBlock(t)

	db.View(func(txn StoreTxn) error {
		uncheckedBlock, err := txn.GetUncheckedBlock(parentHash, kind)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("block not found")
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(uncheckedBlock)
		}
		return nil
	})
}

func TestBadgerStoreTxn_HasUncheckedBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	parentHash, _, kind := parseUncheckedBlock(t)
	db.View(func(txn StoreTxn) error {
		r, err := txn.HasUncheckedBlock(parentHash, kind)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(r)
		return nil
	})
}

func TestBadgerStoreTxn_CountUncheckedBlocks(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	db.View(func(txn StoreTxn) error {
		b, err := txn.CountUncheckedBlocks()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(b)
		return nil
	})
}

func TestBadgerStoreTxn_DeleteUncheckedBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	parentHash, _, kind := parseUncheckedBlock(t)

	db.Update(func(txn StoreTxn) error {
		err := txn.DeleteUncheckedBlock(parentHash, kind)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
}

// Test Badger Representation CURD

func parseRepresentation(t *testing.T) (address types.Address) {
	address, err := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	return
}

func TestBadgerStoreTxn_GetRepresentation(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	address := parseRepresentation(t)
	db.View(func(txn StoreTxn) error {
		balance, err := txn.GetRepresentation(address)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("address not found")
			} else {
				t.Fatal(err)
			}
		}
		fmt.Println(balance)
		return nil
	})
}

func TestBadgerStoreTxn_AddRepresentation(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	address := parseRepresentation(t)
	//amount,err := types.ParseBalanceString("1234.12")
	amount, err := types.ParseBalance("400.004", "Mqlc")
	if err != nil {
		t.Fatal(err)
	}
	db.Update(func(txn StoreTxn) error {
		err = txn.AddRepresentation(address, amount)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("address not found")
			} else {
				t.Fatal(err)
			}
		}
		return nil
	})
}

func TestBadgerStoreTxn_SubRepresentation(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	address := parseRepresentation(t)
	amount, err := types.ParseBalance("1000.04", "Mqlc")
	if err != nil {
		t.Fatal(err)
	}
	db.Update(func(txn StoreTxn) error {
		err = txn.SubRepresentation(address, amount)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				fmt.Println("address not found")
			} else {
				t.Fatal(err)
			}
		}
		return nil
	})
}
