package ledger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/test/mock"
)

func TestLedger_MessageInfo(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	mHash := mock.Hash()
	m := []byte{10, 20, 30, 40}
	if err := l.AddMessageInfo(mHash, m); err != nil {
		t.Fatal(err)
	}
	v, err := l.GetMessageInfo(mHash)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(m, v) {
		t.Fatal("err store")
	}
}

func TestLedger_MessageBlock(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	s := "1801111"
	r := "1800000"
	sender, _ := json.Marshal(s)
	receiver, _ := json.Marshal(r)
	message := mock.Hash()
	block := types.StateBlock{
		Type:     types.Send,
		Sender:   sender,
		Receiver: receiver,
		Message:  message,
	}
	if err := l.AddStateBlock(&block); err != nil {
		t.Fatal(err)
	}
	b, err := l.GetSenderBlocks(sender)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != block.GetHash() {
		t.Fatal("err store")
	}
	b, err = l.GetReceiverBlocks(receiver)
	if err != nil {
		t.Fatal(err)
	}
	if b[0] != block.GetHash() {
		t.Fatal("err store")
	}
	blk, err := l.GetMessageBlock(message)
	if err != nil {
		t.Fatal(err)
	}
	if blk.GetHash() != block.GetHash() {
		t.Fatal("err store")
	}
}
