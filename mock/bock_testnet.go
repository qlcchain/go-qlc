// +build  testnet

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
			 "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
    		 "address": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
    		 "balance": "0",
   			 "vote": "0",
   			 "network": "0",
   			 "storage": "0",
   			 "oracle": "0",
   			 "previous": "5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d",
   			 "link": "6c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7",
   			 "sender": "MTU4MTExMTAwMDAw",
   			 "receiver": "MTg1MDAwMDExMTE=",
   			 "message": "e4d85318e2898ee2ef3ac5baa3c6234a45cc84898756b64c08c3df48adce6492",
   			 "povHeight": 0,
   			 "timestamp": 1555747957,
   			 "extra": "0000000000000000000000000000000000000000000000000000000000000000",
   			 "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
   			 "work": "00000000002b1b77",
   			 "signature": "5085a4a1a9ff5cd46f9891aebdc623faf9fd533dacc00c640ad05bebf9081b120fd5ea625bd9030ecb5e6ce836af29feaf85365354165c877a6335d53767e00a"
        }`
	jsonTestReceive = `{
        	 "type": "Open",
             "token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
   			 "address": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
   			 "balance": "60000000000000000",
   			 "vote": "0",
   			 "network": "0",
   			 "storage": "0",
   			 "oracle": "0",
   			 "previous": "0000000000000000000000000000000000000000000000000000000000000000",
   			 "link": "6225d59d67b03594fa0fa37f02c9a34cd5fea2cb04c4a9ee7634130775ad5dd7",
   			 "message": "0000000000000000000000000000000000000000000000000000000000000000",
   			 "povHeight": 0,
   			 "timestamp": 1555748290,
   			 "extra": "0000000000000000000000000000000000000000000000000000000000000000",
   			 "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
             "work": "00000000006a915e",
             "signature": "4555adb11ec66cf5dcf06925cb3aab4aaba7f79dd34f91f38f7e23835a552963946155cb272fc25831b65211799ad680a1b899a5887d9f83a0d3ef4a6e88fa0e"
        }`
	jsonTestGasSend = `{
        	 "type": "Send",
             "token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
             "address": "qlc_3t1mwnf8u4oyn7wc7wuptnsfz83wsbrubs8hdhgkty56xrrez4x7fcttk5f3",
             "balance": "0",
             "vote": "0",
             "network": "0",
             "storage": "0",
             "oracle": "0",
             "previous": "424b367da2e0ff991d3086f599ce26547b80ae948b209f1cb7d63e19231ab213",
             "link": "6c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7",
             "message": "0000000000000000000000000000000000000000000000000000000000000000",
             "povHeight": 0,
             "timestamp": 1559105475,
             "extra": "0000000000000000000000000000000000000000000000000000000000000000",
             "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
             "work": "00000000015d1b90",
             "signature": "a504678c39c9f5abe456de9c3f240479bcebeef2644362d57e37dbef9fd6f08a4d3e5c40b72047e7be692f1f8f29a095f3725e5dc2760157cd6081eadb28560a"
        }`
	jsonTestGasReceive = `{
        	  "type": "Open",
              "token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
              "address": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
              "balance": "10000000000000000",
              "vote": "0",
              "network": "0",
              "storage": "0",
              "oracle": "0",
              "previous": "0000000000000000000000000000000000000000000000000000000000000000",
              "link": "bbffded05d7ca0e9fd94afa5cc21019ad68b8311de3bf374cbed7b15d87486aa",
              "message": "0000000000000000000000000000000000000000000000000000000000000000",
              "povHeight": 0,
              "timestamp": 1559105597,
              "extra": "0000000000000000000000000000000000000000000000000000000000000000",
              "representative": "qlc_3hw8s1zubhxsykfsq5x7kh6eyibas9j3ga86ixd7pnqwes1cmt9mqqrngap4",
              "work": "00000000006a915e",
              "signature": "81dba2c7fd77493181184b0584bd667807321d733b9643a26b17cc3e77b43321590894ce859b5c9e3567cb3e8d38ca456ab32e9adb27e7f564bf7434a8e11108"
        }`

	jsonTestChangeRepresentative = `{
			"type": "Change",
			"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
			"address": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
			"balance": "60000000000000000",
			"vote": "0",
			"network": "0",
			"storage": "0",
			"oracle": "0",
			"previous": "b82a5151d42af6ff1399f2039baf46b8812d28085c4fb5bedadfabea661c242a",
			"link": "0000000000000000000000000000000000000000000000000000000000000000",
			"message": "0000000000000000000000000000000000000000000000000000000000000000",
			"povHeight": 0,
			"timestamp": 1564062913,
			"extra": "0000000000000000000000000000000000000000000000000000000000000000",
			"representative": "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7",
			"work": "0000000000000000",
			"signature": "a9df7a6bb119914f0d22a90b507d5726f9cdf520cd0d82da3c13632b145fe3f3c5e47fe6c8e8124030726d47517ba736b237903d77699a9439dfcf079734c700"
		}`
	MockBlocks = `[
	{
		"type": "Open",
		"token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "1000000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "8b54787c668dddd4f22ad64a8b0d241810871b9a52a989eb97670f345ad5dc90",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "4eaf260b00a0463d9f6c6b097d96d6647afdac1fe119b63e0000edf00540abdf",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "5003c1e7ce87471f1e61cabfa2ab3f9dcf3a37d2bed3a6f1677c832021c7b61c857b0fcaf71293686eb7acf7d57f873f346a7b4f6e9559766084ae609e47050a"
	},
	{
		"type": "Send",
		"token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "8f07e72fb52b02510406f35a934bec012e4c5f219311032cdae505f79194ba95",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "fee77870f1c820138b39a632e7e43cd52e8e7745ef92b69d62979cd41825134d",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "185184711da73fb8784825633cd5d8edb23811a9e457a6876ca0c8e96cdf9f8b14ae62578efbf0dacb86e302a655b2d559257f76cda76951507376b62151610a"
	},
	{
		"type": "Open",
		"token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
		"address": "qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "a8a88cfd9507710bebcaa22a8e3b149bcea01f77eeacbd95c8e6f4729123dad4",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "d4bb5fa089a6c05ad3d7c75f4096d7fbf14cbbffccf9141b9715958102612257",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "742b54922cf764ee5312c60ae5c66bc17c5bfc9e6091a38f03715aedd6ca6ca2bbc2d7f43b7243b0e6996b6ca7059f924f40b6221f99cfdcc0679add768cf900"
	},
	{
		"type": "Open",
		"token": "89066d747a3c74ff1dec8ea6a7011bde010dd404aec454880f23d58cbf9280e4",
		"address": "qlc_1je9h6w3o5b386oig7sb8j71sf6xr9f5ipemw8gojfcqjpk6r5hiu7z3jx3z",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "10f5b4cfa5fc4869c086fb942275c9c82d0f3df0b76ebed14a4a5c1ec643eb01",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "efb7d7f5a253775e1387fa5cef77b38d1a565b75f5f46cd6da344d67d6117ff4",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "125e82660f1433fb147a73c48e820ab716dfdcb8b8caddbf0819afe18ab925d8d2b44c99a7088e2c5c97dc6e83fd0c836535366e53967380d8251f85468b8c01"
	},
	{
		"type": "Open",
		"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "1000000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "5594c690c3618a170a77d2696688f908efec4da2b94363fcb96749516307031d",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "585633676516d1d4cae65163c63d1d064494c3f680172a6289006e7d201dc1ce",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "f700fa4de441d31123f2f7de39792303ff9b2b875a7bec52d0bf8469a91e97f01a68fb043c5e3bdedbdf2603edd1447908724009e23f4fed3031bcfc98f17d08"
	},
	{
		"type": "Send",
		"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
		"address": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "848b09ef7d5a94a64fdb9ad9c0a83de6452963160eec711098a2850082c39742",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "f947765911c8afe004d0e1190a2f8c543b8bcd50aa25e9a80dfad4395cd21bd7",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "3afd1934a28a44c24850d7d40dd37bf0efb6df8e701814a091ee3a36a5e13e45c679b2571dbf9a51f11eba0b13ec449ab360e186e872d3e07eb9bb959a6e8508"
	},
	{
		"type": "Open",
		"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
		"address": "qlc_3pbbee5imrf3aik35ay44phaugkqad5a8qkngot6by7h8pzjrwwmxwket4te",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "71ee369636fbc3f70bf541f663a8706d6c6396558a7a5ae667b739665243cb18",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "755babfebcf48512954184bf47d466626ca3f7051bfa4d8ce46ef9dda7af746a",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "403d5b375e2ef54ef8d6c91e24fae256aae42a3f4f796a775a9e5a4a0e82dfd40c2084fd4b581e901de328a9815c229afe73e04cee77a1abe436e0e6c388ab00"
	},
	{
		"type": "Open",
		"token": "a7e8fa30c063e96a489a47bc43909505bd86735da4a109dca28be936118a8582",
		"address": "qlc_1je9h6w3o5b386oig7sb8j71sf6xr9f5ipemw8gojfcqjpk6r5hiu7z3jx3z",
		"balance": "100000000000000",
		"vote": "0",
		"network": "0",
		"storage": "0",
		"oracle": "0",
		"previous": "0000000000000000000000000000000000000000000000000000000000000000",
		"link": "773b4c091c99aaa6a5a9d22123d06539b64767ef7378f4ff607935e1d9bb5476",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "e8f3b09d46972dc39df45a8f93f7a2fa8e8baeafe28a47929b3c6675e54de642",
		"povHeight": 0,
		"timestamp": 1584356371,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "9aabc16caf0fcd89afdd66283c9e672704ca62a538d7f0e5a11306fab66df953cfb2d5e2a47d7f42753b50c5c6258a1b6fe6ba859d6c765414978faad29d8602"
	}
]`
)
