package mock

import "testing"

func TestMockPov_Generate(t *testing.T) {
	povBlk1, td1 := GeneratePovBlock(nil, 3)
	if povBlk1 == nil {
		t.Fatal()
	}
	if td1 == nil {
		t.Fatal()
	}

	UpdatePovHash(povBlk1)

	auxHdr1 := GenerateAuxPow(povBlk1.GetHash())
	if auxHdr1 == nil {
		t.Fatal()
	}

	povBlk2, td2 := GeneratePovBlockByFakePow(nil, 0)
	if povBlk2 == nil {
		t.Fatal()
	}
	if td2 == nil {
		t.Fatal()
	}
}
