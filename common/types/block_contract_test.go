package types

import (
	"math/rand"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestSmartContractBlock_Serialize(t *testing.T) {
	var sb SmartContractBlock
	account1 := account()
	account2 := account()

	masterAddress := account1.Address()
	i := rand.Intn(100)
	abi := make([]byte, i)
	_ = random.Bytes(abi)
	hash, _ := HashBytes(abi)
	sb.Abi = ContractAbi{Abi: abi, AbiLength: uint64(i), AbiHash: hash}
	sb.Address = account1.Address()
	sb.Owner = account2.Address()

	sb.InternalAccount = masterAddress
	var w Work
	h := sb.GetHash()
	worker, _ := NewWorker(w, h)

	sb.Work = worker.NewWork()
	sb.Signature = account1.Sign(h)

	if hash := sb.GetHash(); hash.IsZero() {
		t.Fatal()
	}

	t.Log(sb.String())
	t.Log(sb.Size())

	buff, err := sb.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	var b2 SmartContractBlock
	if err = b2.Deserialize(buff); err != nil {
		t.Fatal(err)
	}

	if b2.IsUseStorage != sb.IsUseStorage {
		t.Fatal("isUseStorage error")
	}
	if b2.InternalAccount != sb.InternalAccount {
		t.Fatal("internalAccount error")
	}

	if s := sb.IsValid(); !s {
		t.Fatal()
	}
}

func account() *Account {
	seed, _ := NewSeed()
	_, priv, _ := KeypairFromSeed(seed.String(), 0)
	return NewAccount(priv)
}
