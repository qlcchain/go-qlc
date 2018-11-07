package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/qlcchain/go-qlc/common/types"
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
			if blk, err := types.ParseStateBlock(data); err != nil {
				t.Fatal(err)
			} else {
				blocks = append(blocks, blk)
			}
		case types.SmartContract:
			//blk := new(types.SmartContractBlock)
		default:
			t.Fatalf("unsupported block type")
		}
	}
	return
}

func TestBadgerStoreTxn_AddBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	start := time.Now()

	blocks := parseBlocks(t, "../testdata/blocks.json")
	for _, blk := range blocks {
		db.Update(func(txn StoreTxn) error {
			if err := txn.AddBlock(blk); err != nil {
				if err == ErrBlockExists {
					t.Log(err)
				} else {
					t.Fatal(err)
				}
			}
			return nil
		})
	}
	end := time.Now()
	fmt.Printf("write benchmark: time span,%f \n", end.Sub(start).Seconds())
}

func TestBadgerStoreTxn_GetBlock(t *testing.T) {
	db, err := NewBadgerStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	start := time.Now()

	db.View(func(txn StoreTxn) error {
		hash := types.Hash{}
		hash.Of("c22d7a499f40691de472f2f2fb5cff29b660cac4e6b05d412fd324295175a2f1")
		if block, err := txn.GetBlock(hash); err != nil {
			if err == badger.ErrKeyNotFound {
				t.Log(err)
			} else {
				t.Fatal(err)
			}
		} else {
			//fmt.Println(block)
			if block.Hash() != hash {
				t.Fatal(errors.New("get incorrect block"))
			}
		}
		return nil
	})
	end := time.Now()
	fmt.Printf("write benchmark: time span,%f \n", end.Sub(start).Seconds())

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
		hash.Of("bbb23eab1706acaf717be7567c6b4568c801bfbae397503e885f8caf20e968a0")
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
		if block, err := txn.GetRandomBlock(); err != nil {
			if err == ErrStoreEmpty {
				t.Log(err)
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
			if err := txn.AddAccountMeta(accountmeta); err != nil {
				if err == ErrAccountExists {
					t.Log(err)
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
		address, _ := types.HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
		if accountmeta, err := txn.GetAccountMeta(address); err != nil {
			if err == badger.ErrKeyNotFound {
				t.Log(err)
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
		if err = txn.AddTokenMeta(address, &tokenmeta); err != nil {
			if err == ErrTokenExists || err == badger.ErrKeyNotFound {
				t.Log(err)
			} else {
				t.Fatal(err)
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
		if tokenmeta, err := txn.GetTokenMeta(address, tokenType); err != nil {
			if err == ErrTokenNotFound {
				t.Log(err)
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
		if err := txn.AddPending(address, hash, &pendinfo); err != nil {
			if err == ErrPendingExists {
				t.Log(err)
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
		if pendinginfo, err := txn.GetPending(address, hash); err != nil {
			if err == badger.ErrKeyNotFound {
				t.Log(err)
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
		if err := txn.AddUncheckedBlock(parentHash, blk, kind); err != nil {
			if err == ErrBlockExists {
				t.Log(err)
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
		if uncheckedBlock, err := txn.GetUncheckedBlock(parentHash, kind); err != nil {
			if err == badger.ErrKeyNotFound {
				t.Log(err)
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
	address, err := types.HexToAddress("qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i")
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
		if balance, err := txn.GetRepresentation(address); err != nil {
			if err == badger.ErrKeyNotFound {
				t.Log(err)
			} else {
				t.Fatal(err)
			}
		} else {
			fmt.Println(balance)
		}
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
		if err = txn.SubRepresentation(address, amount); err != nil {
			if err == badger.ErrKeyNotFound {
				t.Log(err)
			} else {
				t.Fatal(err)
			}
		}
		return nil
	})
}
