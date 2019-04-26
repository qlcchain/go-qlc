// +build  !testnet

/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package test

import "github.com/qlcchain/go-qlc/common/types"

var (
	testPrivateKey          = "194908c480fddb6e66b56c08f0d55d935681da0b3c9c33077010bf12a91414576c0b2cdd533ee3a21668f199e111f6c8614040e60e70a73ab6c8da036f2a7ad7"
	testAddress             = "qlc_1u1d7mgo8hq5nad8jwesw6azfk53a31ge5minwxdfk8t1fqknypqgk8mi3z7"
	testPledgeSendBlock     types.StateBlock
	testPledgeReceiveBlock  types.StateBlock
	testMintageSendBlock    types.StateBlock
	testMintageReceiveBlock types.StateBlock
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
)
