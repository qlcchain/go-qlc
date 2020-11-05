package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
)

var testBlk = `{
      "type": "send",
      "token":"991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949",
      "address":"qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "balance": "1456778",
      "previous": "247230c7377a661e57d51b17b527198ed52392fb8b99367a234d28ccc378eb05",
      "link": "7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "sender": "IjE1ODExMTEwMDAwMCI=",
      "receiver": "IjE1ODExMTEwMDAwMCI=",
      "message": "1235650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
	  "data": "DCI4Tg==",
	  "quota": 12345612,
	  "timestamp": 783474523,
      "extra": "1235650e78d297037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3",
      "representative": "qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby",
      "work": "3c82cc724905ee00",
      "signature": "5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600"
	}
	`

func TestMarshalStateBlock(t *testing.T) {
	blk := StateBlock{}
	//fmt.Println(blk)
	bytes, err := json.Marshal(blk)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(bytes))
}

func TestUnmarshalStateBlock(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	addr, err := HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
	if err != nil {
		t.Fatal(err)
	}
	if addr != b.Address {
		t.Fatal("addr != address")
	}

	if !b.Balance.Equal(StringToBalance("1456778")) {
		t.Fatal("balance error")
	}
}

func TestStateBlock(t *testing.T) {
	b, err := NewBlock(State)
	if err != nil {
		t.Fatal(err)
	}
	if sb, ok := b.(*StateBlock); ok {
		sb.Balance = Balance{big.NewInt(123454)}
		sb.Address, _ = HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby")
		sb.Token, _ = NewHash("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
		bytes, err := json.Marshal(&sb)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(bytes))
	} else {
		t.Fatal("new state block error")
	}
}

func TestStateBlock_Serialize(t *testing.T) {
	//data := []byte{12, 34, 56, 78}
	//fmt.Println(base64.StdEncoding.EncodeToString(data))

	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.Balance)

	buff, err := b.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(buff))
	var b2 StateBlock
	if err = b2.Deserialize(buff); err != nil {
		t.Fatal(err)
	}
	t.Log(b2)

	blkBytes, _ := json.Marshal(&b2)
	t.Log(string(blkBytes))

	if !b2.Balance.Equal(b.Balance) {
		t.Fatal("balance error")
	}
	if hex.EncodeToString(b2.Data) != hex.EncodeToString(b.Data) {
		t.Fatal("data error")
	}
}

func TestStateBlock_Clone(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}

	val, _ := b.Serialize()
	fmt.Println("length val ", len(val))

	t.Log(b.Balance)
	b.Flag |= BlockFlagSync
	b1 := b.Clone()

	if reflect.DeepEqual(b, b1) {
		t.Fatal("invalid clone")
	}
	if b.String() != b1.String() {
		t.Fatal("invalid clone ", b.String(), b1.String())
	}
	if b.Flag != b1.Flag {
		t.Fatal()
	}
	if b.GetHash() != b1.GetHash() {
		t.Fatal("invalid clone ", b.GetHash(), b1.GetHash())
	}
}

func TestStateBlock_GetData(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}

	if b.GetType().String() != Send.String() {
		t.Fatal()
	}
	if addr, err := HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby"); err != nil || addr != b.GetAddress() {
		t.Fatal()
	}
	if addr, err := HexToAddress("qlc_1c47tsj9cipsda74no7iugu44zjrae4doc8yu3m6qwkrtywnf9z1qa3badby"); err != nil || addr != b.GetRepresentative() {
		t.Fatal()
	}
	if !b.GetBalance().Equal(Balance{Int: big.NewInt(1456778)}) {
		t.Fatal()
	}
	if !b.GetOracle().Equal(ZeroBalance) {
		t.Fatal()
	}
	if !b.GetVote().Equal(ZeroBalance) {
		t.Fatal()
	}
	if !b.GetNetwork().Equal(ZeroBalance) {
		t.Fatal()
	}
	if !b.GetStorage().Equal(ZeroBalance) {
		t.Fatal()
	}
	if h, err := NewHash("247230c7377a661e57d51b17b527198ed52392fb8b99367a234d28ccc378eb05"); err != nil || h != b.GetPrevious() {
		t.Fatal()
	}
	if h, err := NewHash("7d35650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3"); err != nil || h != b.GetLink() {
		t.Fatal()
	}
	if r, err := NewHash("991cf190094c00f0b68e2e5f75f6bee95a2e0bd93ceaa4a6734db9f19b728949"); err != nil || r != b.GetToken() {
		t.Fatal()
	}
	if r, err := NewHash("1235650e78d297037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3"); err != nil || r != b.GetExtra() {
		t.Fatal()
	}
	if r, err := NewHash("1235650e78d8d7037c90390357f8a59bf17eff82cbc03c94f0b6267335a8dcb3"); err != nil || r != b.GetMessage() {
		t.Fatal()
	}
	if r, err := NewSignature("5b11b17db9c8fe0cc58cac6a6eecef9cb122da8a81c6d3db1b5ee3ab065aa8f8cb1d6765c8eb91b58530c5ff5987ad95e6d34bb57f44257e20795ee412e61600"); err != nil || r != b.GetSignature() {
		t.Fatal()
	}
	if bytes.EqualFold([]byte("IjE1ODExMTEwMDAwMCI="), b.GetSender()) {
		t.Fatal()
	}
	if bytes.EqualFold([]byte("IjE1ODExMTEwMDAwMCI="), b.GetReceiver()) {
		t.Fatal()
	}
	if bytes.EqualFold([]byte("DCI4Tg=="), b.GetData()[:]) {
		t.Fatal()
	}
	if !b.TotalBalance().Equal(b.GetBalance().Add(b.GetVote())) {
		t.Fatal()
	}
}

func TestStateBlock_IsValid(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}

	if b.IsOpen() {
		t.Fatal()
	}
	if b.Root() != b.GetPrevious() {
		t.Fatal()
	}
	if b.Parent() != b.GetPrevious() {
		t.Fatal()
	}
	if b.Size() <= 0 {
		t.Fatal()
	}
	if b.IsReceiveBlock() {
		t.Fatal()
	}
	if !b.IsSendBlock() {
		t.Fatal()
	}
	if b.IsContractBlock() {
		t.Fatal()
	}

	b.SetFromSync()
	if !b.IsFromSync() {
		t.Fatal()
	}
}

func TestStateBlockList_Serialize(t *testing.T) {
	b := StateBlock{}
	err := json.Unmarshal([]byte(testBlk), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b.Balance)
	bs := StateBlockList{&b}

	buff, err := bs.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(buff))
	b2 := new(StateBlockList)
	if err = b2.Deserialize(buff); err != nil {
		t.Fatal(err)
	}
	t.Log(b2)

	blkBytes, _ := json.Marshal(&b2)
	t.Log(string(blkBytes))
}

func TestStateBlock_PrivateHash(t *testing.T) {
	blk := StateBlock{}
	blk.Balance = NewBalance(rand.Int63())
	blk.Vote = ToBalance(NewBalance(rand.Int63()))

	h1 := blk.GetHashWithoutPrivacy()

	h2 := blk.GetHash()
	if h1 != h2 {
		t.Fatal("public GetHash != GetHashNotUsed")
	}

	blk.PrivateFrom = ("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1")
	blk.PrivateFor = append(blk.PrivateFor, ("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2"))
	blk.PrivateFor = append(blk.PrivateFor, ("fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff3"))
	blk.PrivateGroupID = ("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff10")

	h3 := blk.GetHash()
	if h1 == h3 {
		t.Fatal("private GetHash == GetHashNotUsed")
	}

	pl := make([]byte, 100, 100)
	blk.SetPrivatePayload(pl)
	if !blk.IsPrivate() {
		t.Fatal("tx IsPrivate")
	}
	if !blk.IsRecipient() {
		t.Fatal("tx IsRecipient")
	}
	if blk.GetPrivatePayload() == nil {
		t.Fatal("tx GetPrivatePayload")
	}
}
