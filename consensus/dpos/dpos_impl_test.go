package dpos

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/qlcchain/go-qlc/common/types"

	"github.com/google/uuid"

	"github.com/qlcchain/go-qlc/config"
	"github.com/qlcchain/go-qlc/mock"
)

func getTestDpos() *DPoS {
	dir := filepath.Join(config.QlcTestDataDir(), "transaction", uuid.New().String())
	cm := config.NewCfgManager(dir)
	return NewDPoS(cm.ConfigFile)
}

func TestGetSeq(t *testing.T) {
	dps := getTestDpos()

	seq1 := dps.getSeq(ackTypeCommon)
	if seq1 != 0 {
		fmt.Printf("expect:0   get:%d", seq1)
		//		t.Fail()
	}

	seq2 := dps.getSeq(ackTypeCommon)
	if seq2 != 1 {
		fmt.Printf("expect:1   get:%d", seq2)
		//		t.Fail()
	}

	seq3 := dps.getSeq(ackTypeFindRep)
	if seq3 != 0x10000002 {
		fmt.Printf("expect:%0X   get:%0X", 0x10000002, seq3)
		//		t.Fail()
	}

	seq4 := dps.getSeq(ackTypeFindRep)
	if seq4 != 0x10000003 {
		fmt.Printf("expect:%0X   get:%0X", 0x10000003, seq4)
		//		t.Fail()
	}
}

func TestGetAckType(t *testing.T) {
	dps := getTestDpos()

	type1 := dps.getAckType(0x10000003)
	if type1 != ackTypeFindRep {
		t.Errorf("expect:%d   get:%d", ackTypeFindRep, type1)
		t.Fail()
	}

	type2 := dps.getAckType(3)
	if type2 != ackTypeCommon {
		t.Errorf("expect:%d   get:%d", ackTypeCommon, type2)
		t.Fail()
	}
}

func TestOnFrontierConfirmed(t *testing.T) {
	dps := getTestDpos()
	block := mock.StateBlockWithoutWork()
	hash := block.GetHash()
	dps.frontiersStatus.Store(hash, frontierChainConfirmed)
	var confirmed bool
	dps.onFrontierConfirmed(hash, &confirmed)

	if !confirmed {
		t.Fatal()
	}
}

func TestBatchVote(t *testing.T) {
	acc := mock.Account()

	hashes := make([]types.Hash, 0)
	hashBytes := make([]byte, 0)

	for i := 0; i < 1024; i++ {
		h := types.HashData([]byte(fmt.Sprintf("%d", i)))
		hashes = append(hashes, h)
		hashBytes = append(hashBytes, h[:]...)
	}

	hash, _ := types.HashBytes(hashBytes)
	sign1 := acc.Sign(hash)

	hashBytes2 := make([]byte, 0)
	for _, h := range hashes {
		hashBytes2 = append(hashBytes2, h[:]...)
	}

	signHash, _ := types.HashBytes(hashBytes2)
	valid := acc.Address().Verify(signHash[:], sign1[:])
	if !valid {
		t.Fatal("sign err")
	}
}
