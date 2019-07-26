// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package test

import "github.com/qlcchain/go-qlc/common/types"

const (
	testPrivateKey = "194908c480fddb6e66b56c08f0d55d935681da0b3c9c33077010bf12a91414576c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7"
	testAddress    = "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7"
)

var (
	testSendBlock            types.StateBlock
	testReceiveBlock         types.StateBlock
	testSendGasBlock         types.StateBlock
	testReceiveGasBlock      types.StateBlock
	testChangeRepresentative types.StateBlock
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
)
