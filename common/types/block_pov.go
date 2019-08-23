package types

import (
	"bytes"
	"encoding/json"
	"github.com/qlcchain/go-qlc/common/util"
	"math/big"
)

//go:generate msgp
//msgp:ignore PovTransactions
//msgp:ignore PovTxByHash

type PovBlockFrom uint16

const (
	PovBlockFromUnkonwn PovBlockFrom = iota
	PovBlockFromLocal
	PovBlockFromRemoteBroadcast
	PovBlockFromRemoteFetch
	PovBlockFromRemoteSync
)

type PovBaseHeader struct {
	Version    uint32 `msg:"v" json:"version"`
	Previous   Hash   `msg:"p,extension" json:"previous"`
	MerkleRoot Hash   `msg:"mr,extension" json:"merkleRoot"`
	Timestamp  uint32 `msg:"ts" json:"timestamp"`
	Bits       uint32 `msg:"b" json:"bits"`
	Nonce      uint32 `msg:"n" json:"nonce"`

	// just for internal use
	Hash   Hash   `msg:"ha,extension" json:"hash"`
	Height uint64 `msg:"he" json:"height"`
}

func (bh *PovBaseHeader) Serialize() ([]byte, error) {
	return bh.MarshalMsg(nil)
}

func (bh *PovBaseHeader) Deserialize(text []byte) error {
	_, err := bh.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovAuxHeader struct {
	AuxMerkleBranch   []Hash       `msg:"amb" json:"auxMerkleBranch"`
	AuxMerkleIndex    int          `msg:"ami" json:"auxMerkleIndex"`
	ParCoinBaseTx     PovBtcTx     `msg:"pcbtx" json:"parCoinBaseTx"`
	ParCoinBaseMerkle []Hash       `msg:"pcbm,extension" json:"parCoinBaseMerkle"`
	ParMerkleIndex    int          `msg:"pmi" json:"parMerkleIndex"`
	ParBlockHeader    PovBtcHeader `msg:"pbh" json:"parBlockHeader"`
	ParentHash        Hash         `msg:"ph,extension" json:"parentHash"`
}

func (ah *PovAuxHeader) Serialize() ([]byte, error) {
	return ah.MarshalMsg(nil)
}

func (ah *PovAuxHeader) Deserialize(text []byte) error {
	_, err := ah.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovCoinBaseTx struct {
	TxNum     uint32  `msg:"tn" json:"txNum"`
	StateHash Hash    `msg:"sh,extension" json:"stateHash"`
	Reward    Balance `msg:"r,extension" json:"reward"`
	CoinBase  Address `msg:"cb,extension" json:"coinBase"`
	Extra     []byte  `msg:"e" json:"extra"`

	// !!! NOT IN HASH !!!
	Signature Signature `msg:"sig,extension" json:"signature"`

	// just for internal use
	Hash Hash `msg:"h,extension" json:"hash"`
}

func (cbtx *PovCoinBaseTx) Serialize() ([]byte, error) {
	return cbtx.MarshalMsg(nil)
}

func (cbtx *PovCoinBaseTx) Deserialize(text []byte) error {
	_, err := cbtx.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (cbtx *PovCoinBaseTx) ComputeHash() Hash {
	buf := new(bytes.Buffer)

	buf.Write(util.LE_Uint32ToBytes(cbtx.TxNum))
	buf.Write(cbtx.StateHash.Bytes())
	buf.Write(cbtx.Reward.Bytes())
	buf.Write(cbtx.CoinBase.Bytes())
	buf.Write(cbtx.Extra)

	txHash := Sha256D_HashData(buf.Bytes())
	return txHash
}

func (cbtx *PovCoinBaseTx) GetHash() Hash {
	if cbtx.Hash.IsZero() {
		cbtx.Hash = cbtx.ComputeHash()
	}

	return cbtx.Hash
}

func (cbtx *PovCoinBaseTx) GetCoinbase() Address {
	return cbtx.CoinBase
}

type PovHeader struct {
	BasHdr PovBaseHeader `msg:"basHdr" json:"basHdr"`
	AuxHdr *PovAuxHeader `msg:"auxHdr" json:"auxHdr"`
	CbTx   PovCoinBaseTx `msg:"cbtx" json:"cbtx"`
}

func (h *PovHeader) GetHash() Hash {
	return h.BasHdr.Hash
}

func (h *PovHeader) GetHeight() uint64 {
	return h.BasHdr.Height
}

func (h *PovHeader) GetPrevious() Hash {
	return h.BasHdr.Previous
}

func (h *PovHeader) GetTimestamp() uint32 {
	return h.BasHdr.Timestamp
}

func (h *PovHeader) GetBits() uint32 {
	return h.BasHdr.Bits
}

func (h *PovHeader) GetTargetInt() *big.Int {
	return CompactToBig(h.BasHdr.Bits)
}

func (h *PovHeader) GetNonce() uint32 {
	return h.BasHdr.Nonce
}

func (h *PovHeader) GetStateHash() Hash {
	return h.CbTx.StateHash
}

func (h *PovHeader) GetCoinBase() Address {
	return h.CbTx.CoinBase
}

func (h *PovHeader) GetReward() Balance {
	return h.CbTx.Reward
}

func (h *PovHeader) Serialize() ([]byte, error) {
	return h.MarshalMsg(nil)
}

func (h *PovHeader) Deserialize(text []byte) error {
	_, err := h.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (h *PovHeader) Copy() *PovHeader {
	newHeader := *h
	return &newHeader
}

func (h *PovHeader) GetAlgoEfficiency() uint {
	algo := PovAlgoType(h.BasHdr.Version & uint32(ALGO_VERSION_MASK))
	switch algo {
	case ALGO_SHA256D:
		return 1
	case ALGO_SCRYPT:
		return 12984
	case ALGO_NIST5:
		return 513
	case ALGO_LYRA2Z:
		return 1973648
	case ALGO_X11:
		return 513
	case ALGO_X16R:
		return 257849
	default:
		return 1 // TODO: we should not be here
	}

	return 1 // TODO: we should not be here
}

func (h *PovHeader) BuildHashData() []byte {
	buf := new(bytes.Buffer)

	buf.Write(util.LE_Uint32ToBytes(h.BasHdr.Version))
	buf.Write(h.BasHdr.Previous.Bytes())
	buf.Write(h.BasHdr.MerkleRoot.Bytes())
	buf.Write(util.LE_Uint32ToBytes(h.BasHdr.Timestamp))
	buf.Write(util.LE_Uint32ToBytes(h.BasHdr.Bits))
	buf.Write(util.LE_Uint32ToBytes(h.BasHdr.Nonce))

	return buf.Bytes()
}

func (h *PovHeader) ComputePowHash() Hash {
	d := h.BuildHashData()

	algo := PovAlgoType(h.BasHdr.Version & uint32(ALGO_VERSION_MASK))
	switch algo {
	case ALGO_SHA256D:
		powHash := Sha256D_HashData(d)
		return powHash
	case ALGO_SCRYPT:
		powHash := Scrypt_HashData(d)
		return powHash
	}

	return FFFFHash
}

func (h *PovHeader) ComputeHash() Hash {
	hash, _ := Sha256D_HashBytes(h.BuildHashData())
	return hash
}

type PovBody struct {
	Txs []*PovTransaction `msg:"txs" json:"txs"`
}

func (b *PovBody) Serialize() ([]byte, error) {
	return b.MarshalMsg(nil)
}

func (b *PovBody) Deserialize(text []byte) error {
	_, err := b.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (b *PovBody) Copy() *PovBody {
	copyBd := *b
	return &copyBd
}

type PovStoreBody struct {
	Txs []*PovTransaction `msg:"txs" json:"txs"`
}

func (b *PovStoreBody) Serialize() ([]byte, error) {
	return b.MarshalMsg(nil)
}

func (b *PovStoreBody) Deserialize(text []byte) error {
	_, err := b.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovBlock struct {
	Header PovHeader `msg:"h" json:"header"`
	Body   PovBody   `msg:"b" json:"body"`
}

func NewPovBlockWithBody(header *PovHeader, body *PovBody) *PovBlock {
	copyHdr := header.Copy()
	copyBd := body.Copy()

	blk := &PovBlock{
		Header: *copyHdr,
		Body:   *copyBd,
	}

	return blk
}

func (blk *PovBlock) GetHeader() *PovHeader {
	return &blk.Header
}

func (blk *PovBlock) GetBody() *PovBody {
	return &blk.Body
}

func (blk *PovBlock) GetHash() Hash {
	return blk.Header.BasHdr.Hash
}

func (blk *PovBlock) GetHeight() uint64 {
	return blk.Header.BasHdr.Height
}

func (blk *PovBlock) GetPrevious() Hash {
	return blk.Header.BasHdr.Previous
}

func (blk *PovBlock) GetMerkleRoot() Hash {
	return blk.Header.BasHdr.MerkleRoot
}

func (blk *PovBlock) GetTimestamp() uint32 {
	return blk.Header.BasHdr.Timestamp
}

func (blk *PovBlock) GetTargetInt() *big.Int {
	return blk.Header.GetTargetInt()
}

func (blk *PovBlock) GetStateHash() Hash {
	return blk.Header.CbTx.StateHash
}

func (blk *PovBlock) GetCoinBase() Address {
	return blk.Header.CbTx.CoinBase
}

func (blk *PovBlock) GetSignature() Signature {
	return blk.Header.CbTx.Signature
}

func (blk *PovBlock) GetTxNum() uint32 {
	return blk.Header.CbTx.TxNum
}

func (blk *PovBlock) GetAllTxs() []*PovTransaction {
	return blk.Body.Txs
}

func (blk *PovBlock) GetCoinBaseTx() *PovTransaction {
	return blk.Body.Txs[0]
}

func (blk *PovBlock) GetAccountTxs() []*PovTransaction {
	return blk.Body.Txs[1:]
}

func (blk *PovBlock) GetBits() uint32 {
	return blk.Header.BasHdr.Bits
}

func (blk *PovBlock) ComputeHash() Hash {
	return blk.Header.ComputeHash()
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
	Hash  Hash           `msg:"h,extension" json:"hash"`
	CbTx  *PovCoinBaseTx `msg:"-" json:"-"`
	Block *StateBlock    `msg:"-" json:"-"`
}

func (tx *PovTransaction) GetHash() Hash {
	return tx.Hash
}

func (tx *PovTransaction) IsCbTx() bool {
	if tx.CbTx != nil {
		return true
	}
	return false
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
	BlockHash   Hash   `msg:"bha,extension" json:"blockHash"`
	BlockHeight uint64 `msg:"bhe" json:"blockHeight"`
	TxIndex     uint64 `msg:"ti" json:"txIndex"`
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

type PovMinerStatItem struct {
	FirstHeight  uint64  `msg:"fh" json:"firstHeight"`
	LastHeight   uint64  `msg:"lh" json:"lastHeight"`
	BlockNum     uint32  `msg:"bn" json:"blockNum"`
	RewardAmount Balance `msg:"ra,extension" json:"rewardAmount"`
}

func (msi *PovMinerStatItem) Serialize() ([]byte, error) {
	return msi.MarshalMsg(nil)
}

func (msi *PovMinerStatItem) Deserialize(text []byte) error {
	_, err := msi.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovMinerDayStat struct {
	DayIndex uint32 `msg:"di" json:"dayIndex"`
	MinerNum uint32 `msg:"mn" json:"minerNum"`

	MinerStats map[string]*PovMinerStatItem `msg:"mss" json:"minerStats"`
}

func (mds *PovMinerDayStat) Serialize() ([]byte, error) {
	return mds.MarshalMsg(nil)
}

func (mds *PovMinerDayStat) Deserialize(text []byte) error {
	_, err := mds.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func NewPovMinerDayStat() *PovMinerDayStat {
	ds := new(PovMinerDayStat)
	ds.MinerStats = make(map[string]*PovMinerStatItem)
	return ds
}

type PovTD struct {
	Chain   BigNum `msg:"c,extension" json:"chain"`
	Sha256d BigNum `msg:"sha,extension" json:"sha256d"`
	Scrypt  BigNum `msg:"scr,extension" json:"scrypt"`
	X11     BigNum `msg:"x11,extension" json:"x11"`
}

func (td *PovTD) Serialize() ([]byte, error) {
	return td.MarshalMsg(nil)
}

func (td *PovTD) Deserialize(text []byte) error {
	_, err := td.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (td *PovTD) Copy() *PovTD {
	copyTD := new(PovTD)
	copyTD.Chain = *td.Chain.Copy()
	copyTD.Sha256d = *td.Sha256d.Copy()
	copyTD.Scrypt = *td.Scrypt.Copy()
	copyTD.X11 = *td.X11.Copy()
	return copyTD
}

func NewPovTD() *PovTD {
	td := new(PovTD)
	return td
}
