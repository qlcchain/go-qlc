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
	block := types.StateBlock{
		Type:     types.Send,
		Sender:   sender,
		Receiver: receiver,
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
	if err := l.DeleteStateBlock(block.GetHash()); err != nil {
		t.Fatal(err)
	}
	b, err = l.GetReceiverBlocks(receiver)
	if err != nil || len(b) != 0 {
		t.Fatal(err)
	}

}

func TestLedger_GetMessageBlocks(t *testing.T) {
	teardownTestCase, l := setupTestCase(t)
	defer teardownTestCase(t)

	message := mock.Hash()
	block1 := types.StateBlock{
		Type:    types.Send,
		Message: message,
	}
	block2 := types.StateBlock{
		Type:    types.Receive,
		Message: message,
	}
	if err := l.AddMessageInfo(message, []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(&block1); err != nil {
		t.Fatal(err)
	}
	if err := l.AddStateBlock(&block2); err != nil {
		t.Fatal(err)
	}
	h, err := l.GetMessageBlocks(message)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 2 {
		t.Fatal()
	}
	if err := l.DeleteStateBlock(block1.GetHash()); err != nil {
		t.Fatal(err)
	}
	h, err = l.GetMessageBlocks(message)
	if err != nil || len(h) != 1 {
		t.Fatal(err)
	}
	if _, err := l.GetMessageInfo(message); err != nil {
		t.Fatal(err)
	}
	if err := l.DeleteStateBlock(block2.GetHash()); err != nil {
		t.Fatal(err)
	}
	h, err = l.GetMessageBlocks(message)
	if err != nil || len(h) != 0 {
		t.Fatal(err)
	}
	if _, err := l.GetMessageInfo(message); err == nil {
		t.Fatal(err)
	}
}
