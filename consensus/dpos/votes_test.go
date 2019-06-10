package dpos

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/p2p/protos"
)

var (
	testVotesBlk = `{
    "type": "state",
	"addresses": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
	"representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"balance": "00000000000000000000000000000060",
	"link": "D5BA6C7BB3F4F6545E08B03D6DA1258840E0395080378A890601991A2A9E3163",
	"token": "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
	"signature": "AD57AA8819FA6A7811A13FF0684A79AFDEFB05077BCAD4EC7365C32D2A88D78C8C7C54717B40C0888A0692D05BF3771DF6D16A1F24AE612172922BBD4D93370F",
	"work": "13389dd67ced8429"
	}
	`
	testVotesBlk1 = `{
    "type": "state",
	"addresses": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"previous": "0000000000000000000000000000000000000000000000000000000000000000",
	"representative": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
	"balance": "00000000000000000000000000000050",
	"link": "D5BA6C7BB3F4F6545E08B03D6DA1258840E0395080378A890601991A2A9E3163",
	"token": "125998E086F7011384F89554676B69FCD86769642080CE7EED4A8AA83EF58F36",
	"signature": "AD57AA8819FA6A7811A13FF0684A79AFDEFB05077BCAD4EC7365C32D2A88D78C8C7C54717B40C0888A0692D05BF3771DF6D16A1F24AE612172922BBD4D93370F",
	"work": "13389dd67ced8429"
	}
	`
)

func TestVotes(t *testing.T) {
	blk := new(types.StateBlock)
	if err := json.Unmarshal([]byte(testVotesBlk), &blk); err != nil {
		t.Fatal("Unmarshal block error")
	}
	blk1 := new(types.StateBlock)
	if err := json.Unmarshal([]byte(testVotesBlk1), &blk1); err != nil {
		t.Fatal("Unmarshal block error")
	}
	vts := NewVotes(blk)
	exit, _ := vts.voteExit(blk.GetAddress())
	if exit != false {
		t.Fatal("vote exit func error")
	}
	var seedstring = "DB68096C0E2D2954F59DA5DAAE112B7B6F72BE35FC96327FE0D81FD0CE5794A9"
	s, err := hex.DecodeString(seedstring)
	if err != nil {
		t.Fatal("hex string error")
	}
	seed, err := types.BytesToSeed(s)
	if err != nil {
		t.Fatal("bytes to seed error")
	}
	ac, err := seed.Account(0)
	if err != nil {
		t.Fatal("seed to account error")
	}
	var va protos.ConfirmAckBlock
	va.Sequence = 0
	va.Blk = blk
	va.Account = ac.Address()
	va.Signature = ac.Sign(blk.GetHash())
	status := vts.voteStatus(&va)
	if status != vote {
		t.Fatal("vote status error: vote")
	}
	status = vts.voteStatus(&va)
	if status != confirm {
		t.Fatal("vote status error: confirm")
	}
	var vb protos.ConfirmAckBlock
	vb.Sequence = 0
	vb.Blk = blk1
	vb.Account = ac.Address()
	vb.Signature = ac.Sign(blk1.GetHash())
	status = vts.voteStatus(&vb)
	if status != changed {
		t.Fatal("vote status error: changed")
	}
}
