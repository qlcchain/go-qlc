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

	MockBlocks = `[
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "1000000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "c0d330096ec4ab6ccf5481e06cc54e74b14f534e99e38df486f47d1123cbd1ae",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "67db901ccbbc39e67b7c2ed203b0143c705e7b1b3f38c49c841671c1851a8586",
		"povHeight": 0,
		"timestamp": 1584356165,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000122e972",
		"signature": "bb5c2e6d3b30f2edd749669452a447c0dfd45538edc09d32e407c22ba4c728f7945aec3dd253405b2b7eb81543e081a91edccf10362a8bbe722b75021305d901"
	},
	{
		"type": "Send",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "6a32db61ecca1c4e3ecaa025489d98ce13912dc98052f62dbbf0365847b1e20c",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "145f814eaecf4db4ecc746f5842c904191c6282e46c35399dbe03a1d92167e2a",
		"povHeight": 0,
		"timestamp": 1584356169,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000000d33f",
		"signature": "68cd37b21db416689a4664f234737dc96adcf8c4142682253cfd9ce54a5c9adfe46ea1610346d44533ce9f1a0445dc355ada9bef70f4f8b7ef1bff2b63019409"
	},
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "2341d58e98c9838013e268de9b1040ab3255ff690dc29730b9dec003f833af56",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "45e6adc6e4953516e70a5ffba882b77377e5203fc508adfbbde5ecb99819e907",
		"povHeight": 0,
		"timestamp": 1584356169,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002389ad",
		"signature": "3838ef1ba28b68a73e1a4c663734e8452cccb6aa3873ad2156b7f99b19f01748bd50fb2f6987c3a6ead9f757d8429f74cbada8faa6d3d4641ab89d3f6938f50c"
	},
	{
		"type": "Open",
		"token": "ea842234e4dc5b17c33b35f99b5b86111a3af0bd8e4a8822602b866711de6d81",
		"address": "qlc_1je9h6w3o5b386oig7sb8j71sf6xr9f5ipemw8gojfcqjpk6r5hiu7z3jx3z",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "040d3e8181725bbcfed09e8051647a026f8f7d18467bb528934c7f486d368884",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "186e6671cfb150aa25fada7cb27ffdbd7711dc0b277d0a2af237909ae956a302",
		"povHeight": 0,
		"timestamp": 1584356170,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000001e8767a",
		"signature": "2d3b6eec2a835531b613b97a6c2d91e172fd7a1e32972096d74ebcb780da8b0f181a77fed22643f7a97467998db5f852896acc9ec30bd175daae1e1d75fb230e"
	},
	{
		"type": "Open",
		"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "1000000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "7201b4c283b7a32e88ec4c5867198da574de1718eb18c7f95ee8ef733c0b5609",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "48116e9c726e0a09eadd3ba1c35555b7a5f9a106812774b1f8e9578e75ae3ac2",
		"povHeight": 0,
		"timestamp": 1584356177,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000122e972",
		"signature": "4db48600369fc6fb95d9c65670f2ed3fc7e0266ea47fe9247c40f638c37ecd886081b2fc4b5ff5730798f6677c5d4a8c3159b83886c249d313bcef8f5d2ec30e"
	},
	{
		"type": "Send",
		"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "e6ff682faaf3c8cac5f728639ba728bd2be3aa1e8bcf79ce653f168806965152",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "8c3950c8ba813861b92915fefe86a9fef481eeb9d76ee0dfd1418215436cc80f",
		"povHeight": 0,
		"timestamp": 1584356181,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000004f2b48",
		"signature": "ab57c0351311d773dbe16dda4318017d584895bcfbb0b091c37ed06d965d75834c403f85e1d91e6916f10014a8337c35a6bdaef92bacd7682fe625dfd4276102"
	},
	{
		"type": "Open",
		"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
		"address": "qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "f2b8f97e4429491dce54503f4b513f51618f13d6fac207d2aff5f7a4c0324dbd",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "d559990dcbf6254eb30148f88893c7eb072eaa6a6e76c51e8d5b83d2b1e0cd30",
		"povHeight": 0,
		"timestamp": 1584356182,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002389ad",
		"signature": "fbedb60cce1c4a2bfed08b6f14f565fc911c97da731a7b7fcedddaf3ba5b6c9b4f13c2705f2a8e3c3366aed61ac8f781f289723b04a76626bc63afa8ea123702"
	},
	{
		"type": "Open",
		"token": "45dd217cd9ff89f7b64ceda4886cc68dde9dfa47a8a422d165e2ce6f9a834fad",
		"address": "qlc_1je9h6w3o5b386oig7sb8j71sf6xr9f5ipemw8gojfcqjpk6r5hiu7z3jx3z",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "05f9f70efd1b61f219d665b7879d1a904961a22887d3654d6a245046090a8ef3",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "f3c7f231d36832ae6a517a5e91423864425d5c657d358aa3802d84e285177034",
		"povHeight": 0,
		"timestamp": 1584356183,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000001e8767a",
		"signature": "7e00646271232de77b7ec1b4cac8d4039687eb5a4170501987de40047d19bdf4c871b60dfd8ffa810013af3ef75a5cf7fa4e1274f1d423d2b4383380e2d9de0f"
	}
]`
)
