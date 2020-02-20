// +build  !testnet

package mock

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/types"
)

func init() {
	_ = json.Unmarshal([]byte(jsonTestSend), &TestSendBlock)
	_ = json.Unmarshal([]byte(jsonTestReceive), &TestReceiveBlock)
	_ = json.Unmarshal([]byte(jsonTestGasSend), &TestSendGasBlock)
	_ = json.Unmarshal([]byte(jsonTestGasReceive), &TestReceiveGasBlock)
	_ = json.Unmarshal([]byte(jsonTestChangeRepresentative), &TestChangeRepresentative)
}

var (
	TestSendBlock            types.StateBlock
	TestReceiveBlock         types.StateBlock
	TestSendGasBlock         types.StateBlock
	TestReceiveGasBlock      types.StateBlock
	TestChangeRepresentative types.StateBlock
)

var (
	jsonTestSend = `{
        	  "type": "Send",
              "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
              "address": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
   			  "balance": "0",
    		  "vote": "0",
              "network": "0",
              "storage": "0",
              "oracle": "0",
              "previous": "7201b4c283b7a32e88ec4c5867198da574de1718eb18c7f95ee8ef733c0b5609",
              "link": "6c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7",
              "sender": "MTU4MTExMTAwMDAw",
              "receiver": "MTg1MDAwMDExMTE=",
              "message": "e4d85318e2898ee2ef3ac5baa3c6234a45cc84898756b64c08c3df48adce6492",
              "povHeight": 0,
              "timestamp": 1555748661,
              "extra": "0000000000000000000000000000000000000000000000000000000000000000",
              "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
              "work": "0000000000041639",
              "signature": "614b357c1bb5f9ce920f9cad48b07fb641d233d5da039a17165c3ef7c4409898d17b6d186034fe7aa69c7484081c2ec06db35af2b9f93e0928730f67572de800"
        }`
	jsonTestReceive = `{
        	  "type": "Open",
    		  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
              "address": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
              "balance": "60000000000000000",
              "vote": "0",
              "network": "0",
              "storage": "0",
              "oracle": "0",
              "previous": "0000000000000000000000000000000000000000000000000000000000000000",
              "link": "e13161eb56deb0f4c8498e377e4ffea83b994f4253a5b1a6590fa21f65db43e8",
              "message": "0000000000000000000000000000000000000000000000000000000000000000",
              "povHeight": 0,
              "timestamp": 1555748778,
              "extra": "0000000000000000000000000000000000000000000000000000000000000000",
              "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
              "work": "00000000006a915e",
              "signature": "dcf33e2725c1c9622567b4513dd08b7ebeaff9f4516d588d8a11af873e53dc96895ee73b6b54a079ac1aca4445866a5d32ea4d57c658de4c90255845a86a1208"
        }`
	jsonTestGasSend = `{
              "type": "Send",
              "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
              "address": "qlc_3qe19joxq85rnff5wj5ybp6djqtheqqetfgqc3iogxagnjq4rrbmbp1ews7d",
              "balance": "0",
              "vote": "0",
              "network": "0",
              "storage": "0",
              "oracle": "0",
              "previous": "b9e2ea2e4310c38ed82ff492cb83229b4361d89f9c47ebbd6653ddec8a07ebe1",
              "link": "6c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7",
              "message": "0000000000000000000000000000000000000000000000000000000000000000",
              "povHeight": 0,
              "timestamp": 1559105050,
              "extra": "0000000000000000000000000000000000000000000000000000000000000000",
              "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
              "work": "000000000072e440",
              "signature": "67936b1b054222c549320dbdfcae7575eabb6a0f220eb1a90e522cb0dcc92df346b91049f7a5b0ac590e5548dd04b792d6712af9ce13d864abad0aea13605608"
        }`
	jsonTestGasReceive = `{
              "type": "Open",
              "token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
              "address": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
              "balance": "10000000000000000",
              "vote": "0",
              "network": "0",
              "storage": "0",
              "oracle": "0",
              "previous": "0000000000000000000000000000000000000000000000000000000000000000",
              "link": "bde2611ecac4f58f18c244df59ecec880d440c77203438d809a09b4872efbcab",
              "message": "0000000000000000000000000000000000000000000000000000000000000000",
              "povHeight": 0,
              "timestamp": 1559105225,
              "extra": "0000000000000000000000000000000000000000000000000000000000000000",
              "representative": "qlc_1t1uynkmrs597z4ns6ymppwt65baksgdjy1dnw483ubzm97oayyo38ertg44",
              "work": "00000000006a915e",
              "signature": "e870c547b3acfb38493a090d138c6e8d59bf2f6b75642f7024a76da61314a7e62c2a7dd5d50a5eaa29cd9a005aa9ba03ff4e8eb7ffd7484733ee956d1f0ae606"
        }`
	jsonTestChangeRepresentative = `{
              "type": "Change",
   			  "token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
              "address": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
              "balance": "60000000000000000",
              "vote": "0",
              "network": "0",
              "storage": "0",
              "oracle": "0",
              "previous": "f9f777a74dd6a2d768237a03c158de37884fee14ff6c6eb84d637395134a3e2f",
              "link": "0000000000000000000000000000000000000000000000000000000000000000",
              "message": "0000000000000000000000000000000000000000000000000000000000000000",
              "povHeight": 0,
              "timestamp": 1564729656,
              "extra": "0000000000000000000000000000000000000000000000000000000000000000",
              "representative": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
              "work": "00000000006be765",
              "signature": "dcd82745d7fc037243e22b103d6c4aab6eebf9ffe3db2025fbf6514ff8e050651f35959a4239e98cfdcb68a67b7495d9f57aa4ee93d0fecdc1b882aa473b8800"
        }`
)
