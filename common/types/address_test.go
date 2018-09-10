package types

import (
	"encoding/hex"
	"testing"
)

var validAddresses = []string{
	"qlc_38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8",
	"qlc_1awsn43we17c1oshdru4azeqjz9wii41dy8npubm4rg11so7dx3jtqgoeahy",
	"qlc_3arg3asgtigae3xckabaaewkx3bzsh7nwz7jkmjos79ihyaxwphhm6qgjps4",
	"qlc_3pczxuorp48td8645bs3m6c3xotxd3idskrenmi65rbrga5zmkemzhwkaznh",
	"qlc_3hd4ezdgsp15iemx7h81in7xz5tpxi43b6b41zn3qmwiuypankocw3awes5k",
	"qlc_1anrzcuwe64rwxzcco8dkhpyxpi8kd7zsjc1oeimpc3ppca4mrjtwnqposrs",
}

var invalidAddresses = []string{
	"qlc_38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r7",
	"xrc_38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8",
	"qlc38nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8",
	"qlc8nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8",
	"qlc_8nm8t5rimw6h6j7wyokbs8jiygzs7baoha4pqzhfw1k79npyr1km8w6y7r8",
}

func TestAddressToPub(t *testing.T) {
	pub, _ := HexToAddress("qlc_3t6k35gi95xu6tergt6p69ck76ogmitsa8mnijtpxm9fkcm736xtoncuohr3")

	if hex.EncodeToString(pub.Bytes()) != "e89208dd038fbb269987689621d52292ae9c35941a7484756ecced92a65093ba" {
		t.Errorf("Address got wrong public key")
	}

}

func TestValidateAddress(t *testing.T) {
	for _, addr := range validAddresses {
		if !IsValidHexAddress(addr) {
			t.Errorf("Valid address did not validate")
		}
	}

	for _, addr := range invalidAddresses {
		if IsValidHexAddress(addr) {
			t.Errorf("Invalid address was validated")
		}
	}
}

func TestKeypairFromSeed(t *testing.T) {
	seed := "1234567890123456789012345678901234567890123456789012345678901234"

	// Generated from the official RaiBlocks wallet using above seed.
	expected := map[uint32]string{
		0: "qlc_3iwi45me3cgo9aza9wx5f7rder37hw11xtc1ek8psqxw5oxb8cujjad6qp9y",
		1: "qlc_3a9d1h6wt3zp8cqd6dhhgoyizmk1ciemqkrw97ysrphn7anm6xko1wxakaa1",
		2: "qlc_1dz36wby1azyjgh7t9nopjm3k5rduhmntercoz545my9s8nm7gcuthuq9fmq",
		3: "qlc_1fb7kaqaue49kf9w4mb9w3scuxipbdm3ez6ibnri4w8qexzg5f4r7on1dmxb",
		4: "qlc_3h9a64yqueuij1j9odt119r3ymm8n83wyyz7o9u7ram1tgfhsh1zqwjtzid9",
	}

	for i := uint32(0); i < uint32(len(expected)); i++ {
		pub, _, err := KeypairFromSeed(seed, i)
		if err != nil || PubToAddress(pub).String() != expected[i] {
			t.Errorf("Wallet generation from seed created the wrong address")
		}
	}
}

func BenchmarkGenerateAddress(b *testing.B) {
	for n := 0; n < b.N; n++ {
		pub, _, _ := GenerateAddress()
		PubToAddress(pub.Bytes())
	}
}
