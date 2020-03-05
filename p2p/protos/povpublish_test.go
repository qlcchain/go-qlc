package protos

import (
	"testing"

	"github.com/qlcchain/go-qlc/mock"
)

func TestPovPublishBlock(t *testing.T) {
	blk, _ := mock.GeneratePovBlock(nil, 0)
	ppb := &PovPublishBlock{
		Blk: blk,
	}
	data, err := PovPublishBlockToProto(ppb)
	if err != nil {
		t.Fatal(err)
	}
	ppb1, err := PovPublishBlockFromProto(data)
	if err != nil {
		t.Fatal(err)
	}
	if ppb.Blk.GetHash() != ppb1.Blk.GetHash() {
		t.Fatal("test pov publish block error")
	}
}
