package ledger

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/ledger/db"
	"github.com/qlcchain/go-qlc/ledger/genesis"
)

const (
	dir string = "testdatabase"
)

func TestLedger_GetAddressHexStr(t *testing.T) {
	addressstr := "qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby"
	if address, err := types.HexToAddress(addressstr); err != nil {
		t.Fatal(err)
	} else {
		hash := new(types.Hash)
		hash.UnmarshalBinary(address.Bytes())
		t.Log(hash)
	}
}

func InitialiseLedger(t *testing.T, genStr string) (*Ledger, error) {
	store, err := db.NewBadgerStore(dir)
	if err != nil {
		return nil, err
	}
	gen, err := genesis.Get(genStr)
	if err != nil {
		return nil, err
	}
	if ledger, err := NewLedger(store, LedgerOptions{Genesis: *gen}); err != nil {
		return nil, err
	} else {
		return ledger, nil
	}
}

var (
	test_send_block = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "b4badad5bf7aa378c35e92b00003c004ba588c6d0a5907db4b866332697876b4",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "500000",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`

	test_open_block = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "0000000000000000000000000000000000000000000000000000000000000000",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "100000",
      "link":"84533798231c7fb7e78a796f7090c1d26db58eab284115beeb07825b187c1780",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`
	test_send_block2 = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "84533798231c7fb7e78a796f7090c1d26db58eab284115beeb07825b187c1780",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "200000",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`

	test_receiver_block = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "6c6f1c5dc777d3d7d3ee0e08c68219f6bca885f35e25baafe33ad3ed2955be01",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "400000",
      "link":"cdefcf2f09f26ec43ef112a9b68f30cf390f4931126081402607bc6d1223ed8b",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`

	test_change_block = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "cdefcf2f09f26ec43ef112a9b68f30cf390f4931126081402607bc6d1223ed8b",
      "representative":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "balance": "200000",
      "link":"0000000000000000000000000000000000000000000000000000000000000000",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`

	test_MissingLink_send_block = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "4e30daf3879b3736bad1274fb237e4c2a77b0a4f1d7c01fe7355f424e0e43a8d",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "250000",
      "link":"2845d6627542d95a0a2a54b0dbb6217e384304baa8ded8664bf258d7b9469fe0",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`

	test_MissingLink_receive_block = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "1b78e3347c70284f2ead6367d224678361fd71210f4045da5785b3f6d92df0ee",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "350000",
      "link":"d66750ccbb0ff65db134efaaec31d0b123a557df34e7e804d6884447ee589b3c",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "3c82cc724905ee00"
	}
	`

	test_MissingPrevious_send_block1 = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "d66750ccbb0ff65db134efaaec31d0b123a557df34e7e804d6884447ee589b3c",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "50000",
      "link":"2845d6627542d95a0a2a54b0dbb6217e384304baa8ded8664bf258d7b9469fe0",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "0082cc724905ee00"
	}
	`

	test_MissingPrevious_send_block2 = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "7cbefae97608ac54b0142bcae45caf3fa1e6e9ebd354dff8f91e21f53fb2caef",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "20000",
      "link":"2845d6627542d95a0a2a54b0dbb6217e384304baa8ded8664bf258d7b9469fe0",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
      "work": "0082cc724905ee00"
	}
	`
)

func TestLedger_AddBlocks(t *testing.T) {
	ledger, err := InitialiseLedger(t, genesis.Test_genesis_data)
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.db.Close()

	blk, err := types.NewBlock(byte(types.State))
	if blk, err = types.ParseStateBlock([]byte(test_send_block)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_open_block)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_send_block2)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_receiver_block)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_change_block)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddBlock_ErrMissingLink(t *testing.T) {
	ledger, err := InitialiseLedger(t, genesis.Test_genesis_data)
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.db.Close()

	blk, err := types.NewBlock(byte(types.State))

	if blk, err = types.ParseStateBlock([]byte(test_MissingLink_receive_block)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil && err != ErrMissingLink {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_MissingLink_send_block)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
}

func TestLedger_AddBlock_ErrMissingPrevious(t *testing.T) {
	ledger, err := InitialiseLedger(t, genesis.Test_genesis_data)
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.db.Close()

	blk, err := types.NewBlock(byte(types.State))
	if blk, err = types.ParseStateBlock([]byte(test_MissingPrevious_send_block2)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil && err != ErrMissingPrevious {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_MissingPrevious_send_block1)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil {
		t.Fatal(err)
	}
}

func TestLedger2_AddBlock(t *testing.T) {
	ledger, err := InitialiseLedger(t, genesis.Test_genesis_data2)
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.db.Close()
}

var (
	test_send_block_ledger2 = `{
      "type": "state",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "previousHash": "247230c7377a661e57d51b17b527198ed52392fb8b99367a234d28ccc378eb05",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "90000",
      "link":"7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
      "work": "3c82cc724905ee00"
	}
	`

	test_receiver_block_ledger2 = `{
      "type": "state",
      "address":"qlc_1zboen99jp8q1fyb1ga5czwcd8zjhuzr7ky19kch3fj8gettjq7mudwuio6i",
      "previousHash": "0000000000000000000000000000000000000000000000000000000000000000",
      "representative":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "9910000",
      "link":"cc71d22c6f7698de3d6912bcfd6e5182557efa682c478329a3f21789dba9aef7",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600",
      "token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
      "work": "3c82cc724905ee00"
	}
	`
)

func TestLedger2_AddBlock2(t *testing.T) {
	ledger, err := InitialiseLedger(t, genesis.Test_genesis_data2)
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.db.Close()
	blk, err := types.NewBlock(byte(types.State))
	if blk, err = types.ParseStateBlock([]byte(test_send_block_ledger2)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil && err != ErrMissingPrevious {
		t.Fatal(err)
	}
	if blk, err = types.ParseStateBlock([]byte(test_receiver_block_ledger2)); err != nil {
		t.Fatal(err)
	}
	if err := ledger.AddBlock(blk); err != nil && err != ErrMissingPrevious {
		t.Fatal(err)
	}
}
