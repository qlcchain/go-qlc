package types

import (
	"encoding/json"
	"testing"
)

var contract = `
  {
    "internalAccount": "qlc_3oftfjxu9x9pcjh1je3xfpikd441w1wo313qjc6ie1es5aobwed5x4pjojic",
    "contract": {
      "abi": "mcvnzY+zF5mVDjsvknvPfFgRToMQAVI4wivQGRZBwerbUIvfrKD6/suZJWiFVOI5sbTa98bpY9+1cUhE2T9yidxSCpvZ4kkBVBMfcL3OJIqG",
      "abiLength": 81,
      "abiHash": "79dab43dcc97205918b297c3aba6259e3ab1ed7d0779dc78eec6f57e5d6307ce"
    },
    "owner": "qlc_1nawsw4yatupd47p3scd5x5i3s9szbsggxbxmfy56f8jroyu945i5seu1cdd",
    "type": "SmartContract",
	"isUseStorage": false,
	"extraAddress":[],
    "address": "qlc_3watpnwym9i43kbkt35yfp8xnqo7c9ujp3b6udajza71mspjfzpnpdgoydzn",
    "previous": "0000000000000000000000000000000000000000000000000000000000000000",
    "extra": "0000000000000000000000000000000000000000000000000000000000000000",
    "work": "00000000007bb1fe",
    "signature": "0e5a3f6b246c99f20bbe302c4cdfb4639ef3358620a85dc996ba38b0e2ae463d3e4a1184b36abfa63d3d11f56c3242446d381f837f2fb6f7336cf303a5cbca08"
  }
`

func TestSmartContractBlock_Serialize(t *testing.T) {

	b := SmartContractBlock{}
	err := json.Unmarshal([]byte(contract), &b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)

	buff, err := b.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	var b2 SmartContractBlock
	if err = b2.Deserialize(buff); err != nil {
		t.Fatal(err)
	}
	t.Log(len(buff))
	t.Log(b2.String())

	//b2.ExtraAddress = []Address{b.Address, b.Address}
	bytes, _ := json.Marshal(&b2)
	t.Log(string(bytes))
	if b2.IsUseStorage != b.IsUseStorage {
		t.Fatal("isUseStorage error")
	}
	if b2.InternalAccount != b.InternalAccount {
		t.Fatal("internalAccount error")
	}
}
