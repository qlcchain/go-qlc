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

const (
	PovTxVersion                 = 1
	PovMaxTxInSequenceNum uint32 = 0xffffffff
	PovMaxPrevOutIndex    uint32 = 0xffffffff
)

var (
	PovAuxPowChainID     = 1688
	PovAuxPowHeaderMagic = []byte{0xfa, 0xbe, 'm', 'm'}
)

type PovBaseHeader struct {
	Version    uint32 `msg:"v" json:"version"`
	Previous   Hash   `msg:"p,extension" json:"previous"`
	MerkleRoot Hash   `msg:"mr,extension" json:"merkleRoot"`
	Timestamp  uint32 `msg:"ts" json:"timestamp"`
	Bits       uint32 `msg:"b" json:"bits"` // algo bits
	Nonce      uint32 `msg:"n" json:"nonce"`

	// just for internal use
	Hash   Hash   `msg:"ha,extension" json:"hash"`
	Height uint64 `msg:"he" json:"height"`

	// just for cache use
	NormBits      uint32   `msg:"-" json:"-"` // normalized bits
	NormTargetInt *big.Int `msg:"-" json:"-"` // normalized target big int
	AlgoTargetInt *big.Int `msg:"-" json:"-"` //
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
	AuxMerkleBranch   []*Hash      `msg:"amb" json:"auxMerkleBranch"`
	AuxMerkleIndex    int          `msg:"ami" json:"auxMerkleIndex"`
	ParCoinBaseTx     PovBtcTx     `msg:"pcbtx" json:"parCoinBaseTx"`
	ParCoinBaseMerkle []*Hash      `msg:"pcbm,extension" json:"parCoinBaseMerkle"`
	ParMerkleIndex    int          `msg:"pmi" json:"parMerkleIndex"`
	ParBlockHeader    PovBtcHeader `msg:"pbh" json:"parBlockHeader"`
	ParentHash        Hash         `msg:"ph,extension" json:"parentHash"`
}

func NewPovAuxHeader() *PovAuxHeader {
	aux := new(PovAuxHeader)
	return aux
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

func (ah *PovAuxHeader) ComputePowHash(algo PovAlgoType) Hash {
	return ah.ParBlockHeader.ComputePowHash(algo)
}

type PovCoinBaseTxIn struct {
	PrevTxHash Hash     `msg:"pth,extension" json:"prevTxHash"`
	PrevTxIdx  uint32   `msg:"pti" json:"prevTxIdx"`
	Extra      HexBytes `msg:"ext,extension" json:"extra"` // like BTC's script, filled by miner, 0 ~ 100
	Sequence   uint32   `msg:"seq" json:"sequence"`
}

func (ti *PovCoinBaseTxIn) Copy() *PovCoinBaseTxIn {
	ti2 := *ti
	ti2.Extra = make([]byte, len(ti.Extra))
	copy(ti2.Extra, ti.Extra)
	return &ti2
}

func (ti *PovCoinBaseTxIn) Serialize() ([]byte, error) {
	return ti.MarshalMsg(nil)
}

func (ti *PovCoinBaseTxIn) Deserialize(text []byte) error {
	_, err := ti.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovCoinBaseTxOut struct {
	Value   Balance `msg:"v,extension" json:"value"`
	Address Address `msg:"a,extension" json:"address"`
}

func (to *PovCoinBaseTxOut) Copy() *PovCoinBaseTxOut {
	to2 := *to
	to2.Value = to.Value.Copy()
	return &to2
}

func (to *PovCoinBaseTxOut) Serialize() ([]byte, error) {
	return to.MarshalMsg(nil)
}

func (to *PovCoinBaseTxOut) Deserialize(text []byte) error {
	_, err := to.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PovCoinBaseTx struct {
	Version uint32 `msg:"v" json:"version"`

	// TxIn like BTC's PreviousOutPoint
	TxIns []*PovCoinBaseTxIn `msg:"tis" json:"txIns"`

	// TxOut like BTC
	TxOuts []*PovCoinBaseTxOut `msg:"tos" json:"txOuts"`
	//LockTime uint32 `msg:"lt" json:"lockTime"`

	StateHash Hash   `msg:"sh,extension" json:"stateHash"`
	TxNum     uint32 `msg:"tn" json:"txNum"`

	// just for internal use
	Hash Hash `msg:"h,extension" json:"hash"`
}

func NewPovCoinBaseTx(txInNum int, txOutNum int) *PovCoinBaseTx {
	tx := new(PovCoinBaseTx)
	tx.Version = PovTxVersion

	for i := 0; i < txInNum; i++ {
		txIn := new(PovCoinBaseTxIn)
		txIn.PrevTxIdx = PovMaxPrevOutIndex
		txIn.Sequence = PovMaxTxInSequenceNum
		tx.TxIns = append(tx.TxIns, txIn)
	}
	for i := 0; i < txOutNum; i++ {
		txOut := new(PovCoinBaseTxOut)
		tx.TxOuts = append(tx.TxOuts, txOut)
	}
	return tx
}

func (cbtx *PovCoinBaseTx) Copy() *PovCoinBaseTx {
	tx2 := *cbtx

	tx2.TxIns = nil
	for _, ti := range cbtx.TxIns {
		tx2.TxIns = append(tx2.TxIns, ti.Copy())
	}

	tx2.TxOuts = nil
	for _, to := range cbtx.TxOuts {
		tx2.TxOuts = append(tx2.TxOuts, to.Copy())
	}

	return &tx2
}

func (cbtx *PovCoinBaseTx) GetMinerTxOut() *PovCoinBaseTxOut {
	return cbtx.TxOuts[0]
}

func (cbtx *PovCoinBaseTx) GetRepTxOut() *PovCoinBaseTxOut {
	return cbtx.TxOuts[1]
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

func (cbtx *PovCoinBaseTx) GetCoinBaseData1() []byte {
	buf := new(bytes.Buffer)

	// BTC's Tx Header
	buf.Write(util.LE_Uint32ToBytes(cbtx.Version))
	buf.Write(util.LE_EncodeVarInt(1))

	// TxIn like BTC
	ti := cbtx.TxIns[0]

	buf.Write(ti.PrevTxHash.Bytes())
	buf.Write(util.LE_Uint32ToBytes(ti.PrevTxIdx))

	// TxIn Signature Script will append here

	return buf.Bytes()
}

func (cbtx *PovCoinBaseTx) GetCoinBaseData2() []byte {
	buf := new(bytes.Buffer)

	// TxIn Signature Script will prepend here

	// TxIn Sequence like BTC
	ti := cbtx.TxIns[0]
	buf.Write(util.LE_Uint32ToBytes(ti.Sequence))

	// TxOut like BTC
	buf.Write(util.LE_EncodeVarInt(uint64(len(cbtx.TxOuts))))
	for _, cbto := range cbtx.TxOuts {
		buf.Write(cbto.Address.Bytes())
		buf.Write(cbto.Value.Bytes())
	}

	buf.Write(cbtx.StateHash.Bytes())
	buf.Write(util.LE_Uint32ToBytes(cbtx.TxNum))

	return buf.Bytes()
}

func (cbtx *PovCoinBaseTx) BuildHashData() []byte {
	buf := new(bytes.Buffer)

	// hash data = coinbase1 + extra + coinbase2
	// extra = miner info + extra nonce1 + extra nonce2

	buf.Write(cbtx.GetCoinBaseData1())

	ti := cbtx.TxIns[0]
	buf.Write(util.LE_EncodeVarInt(uint64(len(ti.Extra))))
	buf.Write(ti.Extra)

	buf.Write(cbtx.GetCoinBaseData2())

	return buf.Bytes()
}

func (cbtx *PovCoinBaseTx) ComputeHash() Hash {
	data := cbtx.BuildHashData()
	txHash := Sha256D_HashData(data)
	return txHash
}

func (cbtx *PovCoinBaseTx) GetHash() Hash {
	if cbtx.Hash.IsZero() {
		cbtx.Hash = cbtx.ComputeHash()
	}

	return cbtx.Hash
}

type PovHeader struct {
	BasHdr PovBaseHeader  `msg:"basHdr" json:"basHdr"`
	AuxHdr *PovAuxHeader  `msg:"auxHdr" json:"auxHdr"`
	CbTx   *PovCoinBaseTx `msg:"cbtx" json:"cbtx"`
}

func NewPovHeader() *PovHeader {
	h := new(PovHeader)
	h.CbTx = NewPovCoinBaseTx(1, 2)
	return h
}

func (h *PovHeader) GetHash() Hash {
	return h.BasHdr.Hash
}

func (h *PovHeader) GetHeight() uint64 {
	return h.BasHdr.Height
}

func (h *PovHeader) GetVersion() uint32 {
	return h.BasHdr.Version
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

func (h *PovHeader) GetAlgoTargetInt() *big.Int {
	if h.BasHdr.AlgoTargetInt == nil {
		h.BasHdr.AlgoTargetInt = CompactToBig(h.BasHdr.Bits)
	}
	return h.BasHdr.AlgoTargetInt
}

func (h *PovHeader) GetNormBits() uint32 {
	if h.BasHdr.NormBits == 0 {
		h.GetNormTargetInt()
	}
	return h.BasHdr.NormBits
}

func (h *PovHeader) GetNormTargetInt() *big.Int {
	if h.BasHdr.NormTargetInt == nil {
		nt := h.GetAlgoTargetInt()
		h.BasHdr.NormTargetInt = new(big.Int).Div(nt, big.NewInt(int64(h.GetAlgoEfficiency())))
		h.BasHdr.NormBits = BigToCompact(h.BasHdr.NormTargetInt)
	}
	return h.BasHdr.NormTargetInt
}

func (h *PovHeader) GetNonce() uint32 {
	return h.BasHdr.Nonce
}

func (h *PovHeader) GetTxNum() uint32 {
	return h.CbTx.TxNum
}

func (h *PovHeader) GetStateHash() Hash {
	return h.CbTx.StateHash
}

func (h *PovHeader) GetMinerAddr() Address {
	return h.CbTx.TxOuts[0].Address
}

func (h *PovHeader) GetMinerReward() Balance {
	return h.CbTx.TxOuts[0].Value
}

func (h *PovHeader) GetRepAddr() Address {
	if len(h.CbTx.TxOuts) < 2 {
		return ZeroAddress
	}
	return h.CbTx.TxOuts[1].Address
}

func (h *PovHeader) GetRepReward() Balance {
	if len(h.CbTx.TxOuts) < 2 {
		return ZeroBalance
	}
	return h.CbTx.TxOuts[1].Value
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
	h2 := *h
	if h.CbTx != nil {
		h2.CbTx = h.CbTx.Copy()
	}
	return &h2
}

func (h *PovHeader) GetAlgoType() PovAlgoType {
	return PovAlgoType(h.BasHdr.Version & uint32(ALGO_VERSION_MASK))
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
	if h.AuxHdr != nil {
		return h.AuxHdr.ComputePowHash(h.GetAlgoType())
	}

	d := h.BuildHashData()

	algo := h.GetAlgoType()
	switch algo {
	case ALGO_SHA256D:
		powHash := Sha256D_HashData(d)
		return powHash
	case ALGO_SCRYPT:
		powHash := Scrypt_HashData(d)
		return powHash
	case ALGO_X11:
		powHash := X11_HashData(d)
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

func NewPovBody() *PovBody {
	return new(PovBody)
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
	copyBd.Txs = nil
	for _, tx := range b.Txs {
		copyBd.Txs = append(copyBd.Txs, tx.Copy())
	}
	return &copyBd
}

type PovBlock struct {
	Header PovHeader `msg:"h" json:"header"`
	Body   PovBody   `msg:"b" json:"body"`
}

func NewPovBlock() *PovBlock {
	b := new(PovBlock)
	b.Header.CbTx = NewPovCoinBaseTx(1, 2)
	return b
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

func (blk *PovBlock) GetVersion() uint32 {
	return blk.Header.BasHdr.Version
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

func (blk *PovBlock) GetStateHash() Hash {
	return blk.Header.CbTx.StateHash
}

func (blk *PovBlock) GetMinerAddr() Address {
	return blk.Header.GetMinerAddr()
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

func (blk *PovBlock) GetAlgoEfficiency() uint {
	return blk.Header.GetAlgoEfficiency()
}

func (blk *PovBlock) GetAlgoType() PovAlgoType {
	return blk.Header.GetAlgoType()
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

func (blk *PovBlock) Copy() *PovBlock {
	copy := NewPovBlockWithBody(blk.GetHeader(), blk.GetBody())
	return copy
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

func (tx *PovTransaction) Copy() *PovTransaction {
	tx2 := *tx
	return &tx2
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
	RepBlockNum  uint32  `msg:"rn" json:"repBlockNum"`
	RepReward    Balance `msg:"rr,extension" json:"repReward"`
	IsMiner      bool    `msg:"im" json:"isMiner"`
}

func NewPovMinerStatItem() *PovMinerStatItem {
	return &PovMinerStatItem{
		RewardAmount: NewBalance(0),
		RepReward:    NewBalance(0),
	}
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
