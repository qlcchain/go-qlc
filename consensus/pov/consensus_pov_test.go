package pov

import (
	"testing"

	"github.com/qlcchain/go-qlc/common"
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/mock"
)

func TestPovConsensus_Fake(t *testing.T) {
	cr := &mockPovConsensusChainReader{}
	cf := NewPovConsensus(PovConsensusModeFake, cr)

	err := cf.Init()
	if err != nil {
		t.Fatal(err)
	}

	err = cf.Start()
	if err != nil {
		t.Fatal(err)
	}

	blk1, _ := mock.GeneratePovBlock(nil, 0)
	hdr1 := blk1.GetHeader()
	hdr1.BasHdr.Nonce = hdr1.BasHdr.Timestamp

	_ = cf.PrepareHeader(hdr1)

	_ = cf.VerifyHeader(hdr1)

	_ = cf.FinalizeHeader(hdr1)

	err = cf.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestPovConsensus_PowTarget(t *testing.T) {
	cr := &mockPovConsensusChainReader{}
	cf := NewPovConsensus(PovConsensusModePow, cr)

	err := cf.Init()
	if err != nil {
		t.Fatal(err)
	}

	err = cf.Start()
	if err != nil {
		t.Fatal(err)
	}

	var allBlks []*types.PovBlock
	for i := 0; i < common.PovChainTargetCycle; i++ {
		blk, _ := mock.GeneratePovBlock(nil, 0)
		allBlks = append(allBlks, blk)
	}
	lastBlk := allBlks[len(allBlks)-1]
	lastHdr := lastBlk.GetHeader()

	_ = cf.PrepareHeader(lastHdr)

	_ = cf.VerifyHeader(lastHdr)

	_ = cf.FinalizeHeader(lastHdr)

	err = cf.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
