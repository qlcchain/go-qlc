package types

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
type PovBlock struct {
	Version int32 `msg:"version" json:"version"`
	Hash    Hash  `msg:"hash,extension" json:"hash"`

	Previous      Hash      `msg:"previous,extension" json:"previous"`
	MerkleRoot    Hash      `msg:"merkleRoot,extension" json:"merkleRoot"`
	Nonce         int64     `msg:"nonce" json:"nonce"`
	VoteSignature Signature `msg:"voteSignature,extension" json:"voteSignature"`

	Height       int64     `msg:"height" json:"height"`
	Timestamp    int64     `msg:"timestamp" json:"timestamp"`
	NextTarget   Signature `msg:"nextTarget,extension" json:"nextTarget"`
	CoinbaseHash Hash      `msg:"coinbaseHash,extension" json:"coinbaseHash"`
	TxCount      int64     `msg:"txCount" json:"txCount"`

	Signature Signature `msg:"signature,extension" json:"signature"`

	Coinbase     StateBlock       `msg:"coinbase" json:"coinbase"`
	Transactions []PovTransaction `msg:"transactions" json:"transactions"`
}

func (b *PovBlock) ComputeVoteHash() Hash {
	hash, _ := HashBytes(b.Previous[:], b.MerkleRoot[:], util.Int2Bytes(b.Nonce))
	return hash
}

func (b *PovBlock) ComputeHash() Hash {
	hash, _ := HashBytes(
		b.Previous[:], b.MerkleRoot[:], util.Int2Bytes(b.Nonce), b.VoteSignature[:],
		util.Int2Bytes(b.Height),
		util.Int2Bytes(b.Timestamp),
		b.NextTarget[:],
		b.CoinbaseHash[:],
		util.Int2Bytes(b.TxCount))
	return hash
}

func (b *PovBlock) GetHash() Hash {
	return b.Hash
}

func (b *PovBlock) GetPrevious() Hash {
	return b.Previous
}

func (b *PovBlock) GetMerkleRoot() Hash {
	return b.MerkleRoot
}

func (b *PovBlock) GetNonce() int64 {
	return b.Nonce
}

func (b *PovBlock) GetVoteSignature() Signature {
	return b.VoteSignature
}

func (b *PovBlock) GetHeight() int64 {
	return b.Height
}

func (b *PovBlock) GetTimestamp() int64 {
	return b.Timestamp
}

func (b *PovBlock) GetNextTarget() Signature {
	return b.NextTarget
}

func (b *PovBlock) GetCoinbaseHash() Hash {
	return b.CoinbaseHash
}

func (b *PovBlock) GetTxCount() int64 {
	return b.TxCount
}

func (b *PovBlock) GetSignature() Signature {
	return b.Signature
}

func (b *PovBlock) Serialize() ([]byte, error) {
	return b.MarshalMsg(nil)
}

func (b *PovBlock) Deserialize(text []byte) error {
	_, err := b.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (b *PovBlock) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

func (b *PovBlock) Clone() *PovBlock {
	clone := PovBlock{}
	bytes, _ := b.Serialize()
	_ = clone.Deserialize(bytes)
	return &clone
}

////go:generate msgp
//type PovCoinbase struct {
//	Token     Hash      `msg:"token,extension" json:"token"`
//	Address   Address   `msg:"address,extension" json:"address"`
//	Balance   Balance   `msg:"balance,extension" json:"balance"`
//	Previous  Hash      `msg:"previous,extension" json:"previous"`
//	Timestamp int64     `msg:"timestamp" json:"timestamp"`
//	PoVHeight uint64    `msg:"povHeight" json:"povHeight"`
//	Signature Signature `msg:"signature,extension" json:"signature"`
//}
//
//func (b *PovCoinbase) ComputeHash() Hash {
//	hash, _ := HashBytes(
//		b.Token[:],
//		b.Address[:],
//		b.Balance.Bytes(),
//		b.Previous[:],
//		util.Int2Bytes(b.Timestamp),
//		util.Uint64ToBytes(b.PoVHeight))
//	return hash
//}
//
//func (c *PovCoinbase) Serialize() ([]byte, error) {
//	return c.MarshalMsg(nil)
//}
//
//func (c *PovCoinbase) Deserialize(text []byte) error {
//	_, err := c.UnmarshalMsg(text)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//go:generate msgp
type PovTransaction struct {
	Hash Hash `msg:"hash,extension" json:"hash"`
}

func (b *PovTransaction) GetHash() Hash {
	return b.Hash
}
