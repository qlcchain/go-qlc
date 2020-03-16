// +build !testnet

/*
 * Copyright (c) 2020. QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package api

const (
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
		"message": "951e0a0c9cd6cd90ff8d8d2a1f2b8f5cb40bbc13bc722dc8e43a632020d70cb1",
		"povHeight": 0,
		"timestamp": 1583410982,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000122e972",
		"signature": "527e08fb3f9d573462c1c93290bf8fca3ad5059a2f6484a329f5bc1bf5c2be7f1d2c95278bdb0d0d8ad50f73471a8a0d67c6897d1c8a0fe30c61089261b57007"
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
		"previous": "0044e37d2ee13c670eadf22918d4995a8c0c366aceba896c2414291eb36233fe",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "76c8da4946250e741c671123641a9c1818c28ca72f2bd87dd83690033170d59d",
		"povHeight": 0,
		"timestamp": 1583410986,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002feb69",
		"signature": "d50ca8aa02416e967df9dc4dfbba2c2120f7d7c1fa153a64dc9e02cb3f0d64af64888923f137569084f23d98f7ebf9acd4ac2a6a415958d451d5d41c49bf6c0f"
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
		"link": "5384e45aae945fb2fedf9404e2aef680b950eeedd93a2b3acb773f7dcda3a103",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "89de59b9d86d3e7c1cee03c1958240610dae6adb9b14bbf8f9c43c850d77796c",
		"povHeight": 0,
		"timestamp": 1583410987,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002389ad",
		"signature": "a6da1cadab5139131a2dfe1d8af25c506f628d7f769466c2d101a8e49a0171cedb81e61d6ab0780eb9a03886f823f3b953de5b4db656d7bbd9de59ae7ecef70b"
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
		"message": "88cb59c4935849d616fc199e241d9ecd2fdb554855b0d8b8db5192622fadcb41",
		"povHeight": 0,
		"timestamp": 1583410987,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "000000000122e972",
		"signature": "8813604e088854d0a7e532cda08e21bee0ce431bdfac71bc2e085088dc02fc930716cecfda18108758396fe1dadef8fbb86380881e55eea8832555db58b5e600"
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
		"previous": "9553f7a2075db0b67549b7b9b87b9e995566cb6c568c4bc290d06f176ccd570e",
		"link": "d929630709e1a1442411a3c2159e8dba5742c6835e54757444f8af35bf1c7393",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "a16425add73f00bea04e309863f06dbc6a02dc9650b763338d25fedb43877620",
		"povHeight": 0,
		"timestamp": 1583410992,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000005c2db4",
		"signature": "572e9b3b6787ceb64d3e226d58fcb10f3689b96b36637ffaaf1374b9cfe30d3e30fd124b467b1d84bb5306954b48aa6bb0d13a4f93f5e6453ab2f59923b0e20d"
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
		"link": "b0fe218bd592b08c10c454bdb88280980531f41058e3e86b0d7253683e3718eb",
		"sender": "MTU4MTExMTAwMDA=",
		"receiver": "MTU4MDAwMDExMTE=",
		"message": "714a64212ddba9a834bcde553173dd91d112cf782a96f284d49ce686290865e7",
		"povHeight": 0,
		"timestamp": 1583410993,
		"extra": "0000000000000000000000000000000000000000000000000000000000000000",
		"representative": "qlc_3pekn1xq8boq1ihpj8q96wnktxiu8cfbe5syaety3bywyd45rkyhmj8b93kq",
		"work": "00000000002389ad",
		"signature": "397d39a586445518588761312373536e7c896a0118bf86daef8f18459e0f0f51c0aa3a70f10857807f48dbe01d71c2f5eab3de8df6267a3b920b931d37854500"
	}
]
`
)
