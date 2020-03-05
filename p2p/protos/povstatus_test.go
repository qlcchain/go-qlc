package protos

import (
	"testing"

	"github.com/qlcchain/go-qlc/common/types"
)

func TestPovStatus(t *testing.T) {
	var currentHash, genesisHash types.Hash
	err := currentHash.Of("D2F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of currentHash error")
	}
	err = genesisHash.Of("12F6F6A6422000C60C0CB2708B10C8CA664C874EB8501D2E109CB4830EA41D47")
	if err != nil {
		t.Fatal("Of genesisHash error")
	}
	ps := &PovStatus{
		CurrentHeight: 100,
		CurrentTD:     []byte{0x01, 0x02},
		CurrentHash:   currentHash,
		GenesisHash:   genesisHash,
		Timestamp:     1583134402,
	}
	data, err := PovStatusToProto(ps)
	if err != nil {
		t.Fatal(err)
	}
	ps1, err := PovStatusFromProto(data)
	if err != nil {
		t.Fatal(err)
	}
	if ps.CurrentHash != ps1.CurrentHash || ps.CurrentHeight != ps1.CurrentHeight || ps.GenesisHash != ps1.GenesisHash || ps.Timestamp != ps1.Timestamp {
		t.Fatal("pov status error")
	}
}
