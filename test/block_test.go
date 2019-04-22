// +build  testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package test

import "github.com/qlcchain/go-qlc/common/types"

var (
	testPrivateKey   = "194908c480fddb6e66b56c08f0d55d935681da0b3c9c33077010bf12a91414576c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7"
	testAddress      = "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7"
	testSendBlock    types.StateBlock
	testReceiveBlock types.StateBlock
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
)
