package types

import (
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/util"
)

//go:generate msgp
//msgp:ignore PovTransactions
//msgp:ignore PovTxByHash

type PovBlockFrom uint16

const (
	Unkonwn PovBlockFrom = iota
	PovBlockFromLocal
	PovBlockFromRemoteBroadcast
	PovBlockFromRemoteFetch
	PovBlockFromRemoteSync
)

// PovHeader represents a block header in the PoV blockchain.
type PovHeader struct {
	Hash Hash `msg:"hash,extension" json:"hash"`

	Previous      Hash      `msg:"previous,extension" json:"previous"`
	MerkleRoot    Hash      `msg:"merkleRoot,extension" json:"merkleRoot"`
	Nonce         uint64    `msg:"nonce" json:"nonce"`
	VoteSignature Signature `msg:"voteSignature,extension" json:"voteSignature"`

	Height    uint64    `msg:"height" json:"height"`
	Timestamp int64     `msg:"timestamp" json:"timestamp"`
	Target    Signature `msg:"target,extension" json:"target"`
	Coinbase  Address   `msg:"coinbase,extension" json:"coinbase"`
	TxNum     uint32    `msg:"txNum" json:"txNum"`
	StateHash Hash      `msg:"stateHash,extension" json:"stateHash"`

	Signature Signature `msg:"signature,extension" json:"signature"`
}

func (header *PovHeader) GetHash() Hash {
	return header.Hash
}

func (header *PovHeader) GetPrevious() Hash {
	return header.Previous
}

func (header *PovHeader) GetMerkleRoot() Hash {
	return header.MerkleRoot
}

func (header *PovHeader) GetNonce() uint64 {
	return header.Nonce
}

func (header *PovHeader) GetVoteSignature() Signature {
	return header.VoteSignature
}

func (header *PovHeader) GetHeight() uint64 {
	return header.Height
}

func (header *PovHeader) GetTimestamp() int64 {
	return header.Timestamp
}

func (header *PovHeader) GetTarget() Signature {
	return header.Target
}

func (header *PovHeader) GetCoinbase() Address {
	return header.Coinbase
}

func (header *PovHeader) GetTxNum() uint32 {
	return header.TxNum
}

func (header *PovHeader) GetStateHash() Hash {
	return header.StateHash
}

func (header *PovHeader) GetSignature() Signature {
	return header.Signature
}

func (header *PovHeader) Copy() *PovHeader {
	newHeader := *header
	return &newHeader
}

func (header *PovHeader) Serialize() ([]byte, error) {
	return header.MarshalMsg(nil)
}

func (header *PovHeader) Deserialize(text []byte) error {
	_, err := header.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

// PovBody is a simple (mutable, non-safe) data container for storing and moving
// a block's data contents (transactions) together.
type PovBody struct {
	Transactions []*PovTransaction `msg:"transactions" json:"transactions"`
}

func (body *PovBody) Serialize() ([]byte, error) {
	return body.MarshalMsg(nil)
}

func (body *PovBody) Deserialize(text []byte) error {
	_, err := body.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

// PovBlock represents an entire block in the PoV blockchain.
type PovBlock struct {
	Hash Hash `msg:"hash,extension" json:"hash"`

	Previous      Hash      `msg:"previous,extension" json:"previous"`
	MerkleRoot    Hash      `msg:"merkleRoot,extension" json:"merkleRoot"`
	Nonce         uint64    `msg:"nonce" json:"nonce"`
	VoteSignature Signature `msg:"voteSignature,extension" json:"voteSignature"`

	Height    uint64    `msg:"height" json:"height"`
	Timestamp int64     `msg:"timestamp" json:"timestamp"`
	Target    Signature `msg:"target,extension" json:"target"`
	Coinbase  Address   `msg:"coinbase,extension" json:"coinbase"`
	TxNum     uint32    `msg:"txNum" json:"txNum"`
	StateHash Hash      `msg:"stateHash,extension" json:"stateHash"`

	Signature Signature `msg:"signature,extension" json:"signature"`

	Transactions []*PovTransaction `msg:"transactions" json:"transactions"`
}

func NewPovBlockWithHeader(header *PovHeader) *PovBlock {
	blk := &PovBlock{
		Hash: header.Hash,

		Previous:      header.Previous,
		MerkleRoot:    header.MerkleRoot,
		Nonce:         header.Nonce,
		VoteSignature: header.VoteSignature,

		Height:    header.Height,
		Timestamp: header.Timestamp,
		Target:    header.Target,
		Coinbase:  header.Coinbase,
		TxNum:     header.TxNum,
		StateHash: header.StateHash,

		Signature: header.Signature,
	}
	return blk
}

func NewPovBlockWithBody(header *PovHeader, body *PovBody) *PovBlock {
	blk := NewPovBlockWithHeader(header)

	copy(blk.Transactions, body.Transactions)

	return blk
}

func (blk *PovBlock) ComputeVoteHash() Hash {
	hash, _ := HashBytes(blk.Previous[:], blk.MerkleRoot[:], util.BE_Uint64ToBytes(blk.Nonce))
	return hash
}

func (blk *PovBlock) ComputeHash() Hash {
	hash, _ := HashBytes(
		blk.Previous[:], blk.MerkleRoot[:], util.BE_Uint64ToBytes(blk.Nonce), blk.VoteSignature[:],
		util.BE_Uint64ToBytes(blk.Height),
		util.BE_Int2Bytes(blk.Timestamp),
		blk.Target[:],
		blk.Coinbase[:],
		util.BE_Uint32ToBytes(blk.TxNum),
		blk.StateHash[:])
	return hash
}

func (blk *PovBlock) GetHeader() *PovHeader {
	header := &PovHeader{
		Hash: blk.Hash,

		Previous:      blk.Previous,
		MerkleRoot:    blk.MerkleRoot,
		Nonce:         blk.Nonce,
		VoteSignature: blk.VoteSignature,

		Height:    blk.Height,
		Timestamp: blk.Timestamp,
		Target:    blk.Target,
		Coinbase:  blk.Coinbase,
		TxNum:     blk.TxNum,
		StateHash: blk.StateHash,

		Signature: blk.Signature,
	}
	return header
}

func (blk *PovBlock) GetBody() *PovBody {
	body := &PovBody{
		Transactions: blk.Transactions,
	}
	return body
}

func (blk *PovBlock) GetHash() Hash {
	return blk.Hash
}

func (blk *PovBlock) GetPrevious() Hash {
	return blk.Previous
}

func (blk *PovBlock) GetMerkleRoot() Hash {
	return blk.MerkleRoot
}

func (blk *PovBlock) GetNonce() uint64 {
	return blk.Nonce
}

func (blk *PovBlock) GetVoteSignature() Signature {
	return blk.VoteSignature
}

func (blk *PovBlock) GetHeight() uint64 {
	return blk.Height
}

func (blk *PovBlock) GetTimestamp() int64 {
	return blk.Timestamp
}

func (blk *PovBlock) GetTarget() Signature {
	return blk.Target
}

func (blk *PovBlock) GetCoinbase() Address {
	return blk.Coinbase
}

func (blk *PovBlock) GetTxNum() uint32 {
	return blk.TxNum
}

func (blk *PovBlock) GetStateHash() Hash {
	return blk.StateHash
}

func (blk *PovBlock) GetSignature() Signature {
	return blk.Signature
}

func (blk *PovBlock) Serialize() ([]byte, error) {
	return blk.MarshalMsg(nil)
}

func (blk *PovBlock) Deserialize(text []byte) error {
	_, err := blk.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (blk *PovBlock) String() string {
	bytes, _ := json.Marshal(blk)
	return string(bytes)
}

func (blk *PovBlock) Clone() *PovBlock {
	clone := PovBlock{}
	bytes, _ := blk.Serialize()
	_ = clone.Deserialize(bytes)
	return &clone
}

type PovBlocks []*PovBlock

func (bs *PovBlocks) Serialize() ([]byte, error) {
	return bs.MarshalMsg(nil)
}

func (bs *PovBlocks) Deserialize(text []byte) error {
	_, err := bs.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

// PovTransaction represents an state block metadata in the PoV block.
type PovTransaction struct {
	Address Address     `msg:"address,extension" json:"address"`
	Hash    Hash        `msg:"hash,extension" json:"hash"`
	Block   *StateBlock `msg:"-" json:"-"`
}

func (tx *PovTransaction) GetHash() Hash {
	return tx.Hash
}

func (tx *PovTransaction) Serialize() ([]byte, error) {
	return tx.MarshalMsg(nil)
}

func (tx *PovTransaction) Deserialize(text []byte) error {
	_, err := tx.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

// PovTransactions is a PovTransaction slice type for basic sorting.
type PovTransactions []*PovTransaction

// Len returns the length of s.
func (s PovTransactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s PovTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// PovTxByHash implements the sort interface to allow sorting a list of transactions
// by their hash.
type PovTxByHash PovTransactions

func (s PovTxByHash) Len() int           { return len(s) }
func (s PovTxByHash) Less(i, j int) bool { return s[i].Hash.Cmp(s[j].Hash) < 0 }
func (s PovTxByHash) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxLookupEntry is a positional metadata to help looking up the data content of
// a transaction given only its hash.
type PovTxLookup struct {
	BlockHash   Hash   `msg:"blockHash,extension" json:"blockHash"`
	BlockHeight uint64 `msg:"blockHeight" json:"blockHeight"`
	TxIndex     uint64 `msg:"txIndex" json:"txIndex"`
}

func (txl *PovTxLookup) Serialize() ([]byte, error) {
	return txl.MarshalMsg(nil)
}

func (txl *PovTxLookup) Deserialize(text []byte) error {
	_, err := txl.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}
