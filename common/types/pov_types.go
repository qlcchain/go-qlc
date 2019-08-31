package types

type PovMineWork struct {
	WorkHash     Hash    `json:"workHash"`
	Version      uint32  `json:"version"`
	Previous     Hash    `json:"previous"`
	Bits         uint32  `json:"bits"`
	Height       uint64  `json:"height"`
	MinTime      uint32  `json:"minTime"`
	MerkleBranch []*Hash `json:"merkleBranch"`
	CoinBaseData []byte  `json:"coinbaseData"`
}

func NewPovMineWork() *PovMineWork {
	w := new(PovMineWork)
	return w
}

type PovMineResult struct {
	WorkHash      Hash
	BlockHash     Hash
	MerkleRoot    Hash
	Timestamp     uint32
	Nonce         uint32
	CoinbaseExtra []byte
	CoinbaseHash  Hash
	CoinbaseSig   Signature
}

func NewPovMineResult() *PovMineResult {
	r := new(PovMineResult)
	return r
}

type PovMineBlock struct {
	Block  *PovBlock
	Header *PovHeader
	Body   *PovBody

	AllTxHashes []*Hash

	WorkHash       Hash
	MinTime        uint32
	CoinbaseBranch []*Hash
}

func NewPovMineBlock() *PovMineBlock {
	mineBlock := new(PovMineBlock)
	mineBlock.Block = NewPovBlock()
	mineBlock.Header = mineBlock.Block.GetHeader()
	mineBlock.Body = mineBlock.Block.GetBody()
	return mineBlock
}
