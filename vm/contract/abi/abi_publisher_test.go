package abi

import (
	"bytes"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/random"
	"github.com/qlcchain/go-qlc/mock"
	"math/big"
	"testing"
)

func TestPublisherPackAndUnPack(t *testing.T) {
	pk := make([]byte, 32)
	verifiers := make([]types.Address, 0)
	codes := make([]types.Hash, 0)

	acc := mock.Address()
	typ := uint32(1)
	id := mock.Hash()
	_ = random.Bytes(pk)
	fee := big.NewInt(5e8)

	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())
	verifiers = append(verifiers, mock.Address())

	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())
	codes = append(codes, mock.Hash())

	data, err := PublisherABI.PackMethod(MethodNamePublish, acc, typ, id, pk, verifiers, codes, fee)
	if err != nil {
		t.Fatal(err)
	}

	var info PublishInfo
	err = PublisherABI.UnpackMethod(&info, MethodNamePublish, data)
	if err != nil {
		t.Fatal(err)
	}

	if info.Account != acc {
		t.Fatal()
	}

	if info.PID != id {
		t.Fatal()
	}

	if info.PType != typ {
		t.Fatal()
	}

	if !bytes.Equal(info.PubKey, pk) {
		t.Fatal()
	}

	if info.Fee.Cmp(fee) != 0 {
		t.Fatal()
	}

	for i, v := range info.Verifiers {
		if v != verifiers[i] {
			t.Fatal()
		}
	}

	for i, c := range info.Codes {
		if c != codes[i] {
			t.Fatal()
		}
	}
}
