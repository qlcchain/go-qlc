/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

const seed = "5a32b2325437cc10c07e36161fcda24f01ec0038969ecaaa709a133372bf4b94"

func TestAccount_Address(t *testing.T) {
	pub, priv, err := KeypairFromSeed(seed, 1)
	if err != nil {
		t.Fatal(err)
	}
	account := NewAccount(priv)
	t.Log(account.String())
	address := account.Address()
	if address != PubToAddress(pub) {
		t.Fatal("invalid address")
	}
	t.Log(address.String())

	h := Hash{}
	err = random.Bytes(h[:])
	if err != nil {
		t.Fatal(err)
	}

	sign := account.Sign(h)
	t.Log(sign.String())

	if !address.Verify(h[:], sign[:]) {
		t.Fatal("sign failed")
	}

	t.Log(hex.EncodeToString(account.PrivateKey()))
}

func TestNewAccount(t *testing.T) {
	if seed, err := NewSeed(); err != nil {
		t.Fatal(err)
	} else {
		if account, err := seed.Account(1); err != nil {
			t.Fatal(err)
		} else {
			t.Log(account.String())
		}
	}
}

var testAccountMeta = `{
    "account": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
    "balance": "17966873799999699",
    "vote": "0",
    "network": "0",
    "storage": "0",
    "oracle": "0",
    "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
    "tokens": [
      {
        "type": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
        "header": "9f5fecaee7faca0ee389b2e93447ba14c682dd02612503ef6117203d45944a19",
        "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        "open": "5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d",
        "balance": "17966873799999699",
        "account": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
        "modified": 1572522444,
        "blockCount": 233,
        "tokenName": "QLC",
        "pending": "0"
      }
    ]
  }
	`

func TestAccountMeta_Clone(t *testing.T) {
	b := AccountMeta{}
	err := json.Unmarshal([]byte(testAccountMeta), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.String())
	b1 := b.Clone()

	if reflect.DeepEqual(b, b1) {
		t.Fatal("invalid clone")
	}

	if b.String() != b1.String() {
		t.Fatal("invalid clone ", b.String(), b1.String())
	}

	if data, err := b.Serialize(); err != nil {
		t.Fatal(err)
	} else {
		b2 := &AccountMeta{}
		if err := b2.Deserialize(data); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(&b, b2) {
				t.Fatalf("exp: %v, act: %v", &b, b2)
			} else {
				if storage := b.GetStorage(); !storage.IsZero() {
					t.Fatalf("storage %s", storage.String())
				}

				if network := b.GetNetwork(); !network.IsZero() {
					t.Fatalf("network %s", network.String())
				}

				if oracle := b.GetOracle(); !oracle.IsZero() {
					t.Fatalf("oracle %s", oracle.String())
				}
			}
		}
	}
}

func TestAccountMeta_Token(t *testing.T) {
	b := AccountMeta{}
	err := json.Unmarshal([]byte(testAccountMeta), &b)
	if err != nil {
		t.Fatal(err)
	}

	h, err := NewHash("a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582")
	if token := b.Token(h); token == nil {
		t.Fatal()
	}

	weight := b.VoteWeight()
	exp := Balance{Int: big.NewInt(17966873799999699)}
	if weight.Compare(exp) != BalanceCompEqual {
		t.Fatalf("invalid vote weight, exp: %s, act: %s", exp.String(), weight.String())
	}

	balance := b.TotalBalance()
	if balance.Compare(exp) != BalanceCompEqual {
		t.Fatalf("invalid balance, exp: %s, act: %s", exp.String(), weight.String())
	}
}
