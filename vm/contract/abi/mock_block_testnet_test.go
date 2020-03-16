// +build testnet

/*
 * Copyright (c) 2020. QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package abi

const (
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
		"message": "03266f9276690456bf4c0421aa9774a764982a8612f67800778fd3ab8dfd1c82",
		"povHeight": 0,
		"timestamp": 1583411111,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "35c8dfd9fd7f7876c9a8ffb0dd05f22d31c5efa764eb03d6f473d261806229f2987b88779db2e8192235afebc3a5dc2fd3b0a317a7107b8a395194b16dbfa20d"
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
		"previous": "1eb275755c76f01596c4b0e3c49f854df55804f4a081803f3771c5c40839fe26",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "b0e43229af1d0bda1d83fd5c89d0d1c4d5690cb1c44d85480195fbcb2481a69f",
		"povHeight": 0,
		"timestamp": 1583411111,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "296f47867b3e467943f1171d06a9bcba934992af78df58dcaf596b924abff22402577fcbe78133ad71b9dcf84f85c84922bf5340e74f8d3526b0d1bcc693d103"
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
		"link": "33c0df0738943130638f361fef0f1519ab50e3b133e670c6ebd1ec3eb8989a7c",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "6ef861004d552ecb2c6cf4ee86b98a657d3a177eca4f48bf79977ad2770f5873",
		"povHeight": 0,
		"timestamp": 1583411111,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "bde35e9fae5d80819aa30df7f41704d508820b2a1f59aa5a6af6a5b6183fe2f2764de174eb28073e251c12d9deb1192a978362166183d07093bff17d02a78b04"
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
		"message": "38009cf01f8ff7139534cac161450f89366e4ce14457d5688604b569705ebb34",
		"povHeight": 0,
		"timestamp": 1583411111,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "ef6c04bb774387fb13f9898f4887054cf2d5669ad21b2ec7216c0f66649a9be02434c61a16b790f3abb4b169009b7905e305b96567a8b679a4ea792e22946401"
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
		"previous": "0a8a0948c50546e0943e17087dd1066634cab9ec46d39ee65c75c293f24825e6",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "11618ba4a2f2f85ba9dfdf20f89ec69c065dcff55a3f956736187fa04923d1a0",
		"povHeight": 0,
		"timestamp": 1583411111,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "e876e26b40768fb252e6f5d0e658f0eec671610351dc52848dd3aae4f8393ad9d42731bfa9be00dc310d38b94e7054a207db6dc967ab92dbd6599e6d050d4a09"
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
		"link": "fdc72f89d92395157fa8d2cac5a7be011b89a53f3e07858a779062eb1dd31213",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "35c3772fd3a68b69b34791283cfbbcce3a0c7cb1bb11fd3d7c2a8e42100f6256",
		"povHeight": 0,
		"timestamp": 1583411111,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "0000000000000000",
		"signature": "937fb58983f5e49c04748d77b29f37b906ba7160057e3fa536bffcfeb1aaec52df53d205f076c398b8c68152e4dd8a00cbe51c86d9b9ba65ae95418fbd1db100"
	}
]
`
)
