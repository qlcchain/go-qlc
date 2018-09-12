package types

import (
	"encoding"
	"errors"
	"encoding/binary"
	"bytes"
	"fmt"
	"github.com/qlcchain/go-qlc/common/types/internal/util"
)

const (
	idBlockInvalid        byte = iota
	idBlockNotABlock
	idBlockState
	idBlockSmart_Contract
)

var (
	ErrBadBlockType = errors.New("bad block type")
	ErrNotABlock    = errors.New("block type is not_a_block")

	blockNames = map[byte]string{
		idBlockInvalid:        "invalid",
		idBlockNotABlock:      "not_a_block",
		idBlockState:          "state",
		idBlockSmart_Contract: "smart_contract",
	}
)

const (
	preambleSize    = 32
	blockSizeCommon = SignatureSize + WorkSize
	blockSizeState  = blockSizeCommon + HashSize*3 + AddressSize*2 + BalanceSize
)

type Block interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	Hash() Hash
	Root() Hash
	TokenHash() Hash
	Size() int
	ID() byte
	Valid(threshold uint64) bool
}

type StateBlock struct {
	Account        Address   `json:"account"`
	PreviousHash   Hash      `json:"previous"`
	Representative Address   `json:"representative"`
	Balance        Balance   `json:"balance"`
	Link           Hash      `json:"link"`
	Token          Hash      `json:"token"`
	Signature      Signature `json:"signature"`
	Work           Work      `json:"work"`
}

func New(blockType byte) (Block, error) {
	switch blockType {
	case idBlockState:
		return new(StateBlock), nil
	case idBlockNotABlock:
		return nil, ErrNotABlock
	default:
		return nil, ErrBadBlockType
	}
}

func Name(id byte) string {
	return blockNames[id]
}

func marshalCommon(sig Signature, work Work, order binary.ByteOrder) ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(sig[:]); err != nil {
		return nil, err
	}

	if err = binary.Write(buf, order, work); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func unmarshalCommon(data []byte, order binary.ByteOrder, sig *Signature, work *Work) error {
	reader := bytes.NewReader(data)

	var s Signature
	if _, err := reader.Read(s[:]); err != nil {
		return err
	}
	*sig = s

	var w Work
	if err := binary.Read(reader, order, &w); err != nil {
		return err
	}
	*work = w

	if reader.Len() != 0 {
		return fmt.Errorf("bad data length: %d unexpected bytes", reader.Len())
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (b *StateBlock) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	var err error
	if _, err = buf.Write(b.Account[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.PreviousHash[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Representative[:]); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Balance.Bytes(binary.BigEndian)); err != nil {
		return nil, err
	}

	if _, err = buf.Write(b.Link[:]); err != nil {
		return nil, err
	}
	if _, err = buf.Write(b.Token[:]); err != nil {
		return nil, err
	}

	commonBytes, err := marshalCommon(b.Signature, b.Work, binary.BigEndian)
	if err != nil {
		return nil, err
	}
	if _, err = buf.Write(commonBytes); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (b *StateBlock) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var err error
	if _, err = reader.Read(b.Account[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.PreviousHash[:]); err != nil {
		return err
	}

	if _, err = reader.Read(b.Representative[:]); err != nil {
		return err
	}

	balance := make([]byte, BalanceSize)
	if _, err = reader.Read(balance); err != nil {
		return err
	}
	if err = b.Balance.UnmarshalBinary(balance); err != nil {
		return err
	}

	if _, err = reader.Read(b.Link[:]); err != nil {
		return err
	}
	if _, err = reader.Read(b.Token[:]); err != nil {
		return err
	}

	commonBytes := make([]byte, blockSizeCommon)
	if _, err = reader.Read(commonBytes); err != nil {
		return err
	}

	return unmarshalCommon(commonBytes, binary.BigEndian, &b.Signature, &b.Work)
}

func (b *StateBlock) Hash() Hash {
	var preamble [preambleSize]byte
	preamble[len(preamble)-1] = idBlockState
	var result Hash
	hh := util.Hash(32, preamble[:], b.Account[:], b.PreviousHash[:], b.Representative[:], b.Balance.Bytes(binary.BigEndian), b.Link[:], b.Token[:])
	copy(result[:], hh)
	return result
}

func (b *StateBlock) Root() Hash {
	if !b.IsOpen() {
		return b.PreviousHash
	}

	return b.Link
}
func (b *StateBlock) TokenHash() Hash {
	return b.Token
}

func (b *StateBlock) Size() int {
	return blockSizeState
}

func (b *StateBlock) ID() byte {
	return idBlockState
}

func (b *StateBlock) Valid(threshold uint64) bool {
	if !b.IsOpen() {
		return b.Work.Valid(b.PreviousHash, threshold)
	}

	return b.Work.Valid(Hash(b.Account), threshold)
}

func (b *StateBlock) IsOpen() bool {
	return b.PreviousHash.IsZero()
}
