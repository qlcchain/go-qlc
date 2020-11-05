package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qlcchain/go-qlc/common/util"
)

const (
	BlockFlagSync uint64 = 1 << iota
)

//go:generate msgp
type StateBlock struct {
	Type           BlockType `msg:"type" json:"type"`
	Token          Hash      `msg:"token,extension" json:"token"`
	Address        Address   `msg:"address,extension" json:"address"`
	Balance        Balance   `msg:"balance,extension" json:"balance"`
	Vote           *Balance  `msg:"vote,extension" json:"vote,omitempty"`
	Network        *Balance  `msg:"network,extension" json:"network,omitempty"`
	Storage        *Balance  `msg:"storage,extension" json:"storage,omitempty"`
	Oracle         *Balance  `msg:"oracle,extension" json:"oracle,omitempty"`
	Previous       Hash      `msg:"previous,extension" json:"previous"`
	Link           Hash      `msg:"link,extension" json:"link"`
	Sender         []byte    `msg:"sender" json:"sender,omitempty"`
	Receiver       []byte    `msg:"receiver" json:"receiver,omitempty"`
	Message        *Hash     `msg:"message,extension" json:"message,omitempty"`
	Data           []byte    `msg:"data" json:"data,omitempty"`
	PoVHeight      uint64    `msg:"povHeight" json:"povHeight"`
	Timestamp      int64     `msg:"timestamp" json:"timestamp"`
	Extra          *Hash     `msg:"extra,extension" json:"extra,omitempty,omitempty"`
	Representative Address   `msg:"representative,extension" json:"representative"`

	PrivateFrom    string   `msg:"priFrom,omitempty" json:"privateFrom,omitempty"`
	PrivateFor     []string `msg:"priFor,omitempty" json:"privateFor,omitempty"`
	PrivateGroupID string   `msg:"priGid,omitempty" json:"privateGroupID,omitempty"`

	Work      Work      `msg:"work,extension" json:"work"`
	Signature Signature `msg:"signature,extension" json:"signature"`

	// following fields just for cache, not marshaled in db or p2p message
	Flag uint64 `msg:"-" json:"-"`

	PrivateRecvRsp bool   `msg:"-" json:"-"`
	PrivatePayload []byte `msg:"-" json:"-"`
}

func (b *StateBlock) BuildHashData() []byte {
	buf := new(bytes.Buffer)

	// fields for public txs
	buf.WriteByte(byte(b.Type))
	buf.Write(b.Token[:])
	buf.Write(b.Address[:])
	buf.Write(b.Balance.Bytes())
	buf.Write(b.GetVote().Bytes())
	buf.Write(b.GetNetwork().Bytes())
	buf.Write(b.GetStorage().Bytes())
	buf.Write(b.GetOracle().Bytes())
	buf.Write(b.Previous[:])
	buf.Write(b.Link[:])
	buf.Write(b.Sender)
	buf.Write(b.Receiver)
	message := b.GetMessage()
	buf.Write(message[:])
	buf.Write(b.Data)
	buf.Write(util.BE_Int2Bytes(b.Timestamp))
	buf.Write(util.BE_Uint64ToBytes(b.PoVHeight))
	extra := b.GetExtra()
	buf.Write(extra[:])
	buf.Write(b.Representative[:])

	// additional fields for private txs
	if len(b.PrivateFrom) > 0 {
		buf.WriteString(b.PrivateFrom)
	}
	if len(b.PrivateFor) > 0 {
		for _, pf := range b.PrivateFor {
			if len(pf) > 0 {
				buf.WriteString(pf)
			}
		}
	}
	if len(b.PrivateGroupID) > 0 {
		buf.WriteString(b.PrivateGroupID)
	}

	return buf.Bytes()
}

func (b *StateBlock) GetHash() Hash {
	data := b.BuildHashData()
	hash := HashData(data)
	return hash
}

func (b *StateBlock) GetHashWithoutPrivacy() Hash {
	t := []byte{byte(b.Type)}
	extra := b.GetExtra()
	message := b.GetMessage()
	hash, _ := HashBytes(t, b.Token[:], b.Address[:], b.Balance.Bytes(), b.GetVote().Bytes(), b.GetNetwork().Bytes(),
		b.GetStorage().Bytes(), b.GetOracle().Bytes(), b.Previous[:], b.Link[:], b.Sender, b.Receiver, message[:], b.Data,
		util.BE_Int2Bytes(b.Timestamp), util.BE_Uint64ToBytes(b.PoVHeight),
		extra[:], b.Representative[:])
	return hash
}

func (b *StateBlock) GetType() BlockType {
	return b.Type
}

func (b *StateBlock) GetToken() Hash {
	return b.Token
}

func (b *StateBlock) GetAddress() Address {
	return b.Address
}

func (b *StateBlock) GetPrevious() Hash {
	return b.Previous
}

func (b *StateBlock) GetBalance() Balance {
	return b.Balance
}

func (b *StateBlock) GetVote() Balance {
	if b.Vote == nil || b.Vote.Int == nil {
		return ZeroBalance
	}
	return *b.Vote
}

func (b *StateBlock) GetOracle() Balance {
	if b.Oracle == nil || b.Oracle.Int == nil {
		return ZeroBalance
	}
	return *b.Oracle
}

func (b *StateBlock) GetNetwork() Balance {
	if b.Network == nil || b.Network.Int == nil {
		return ZeroBalance
	}
	return *b.Network
}

func (b *StateBlock) GetStorage() Balance {
	if b.Storage == nil || b.Storage.Int == nil {
		return ZeroBalance
	}
	return *b.Storage
}

func (b *StateBlock) GetData() []byte {
	return b.Data
}

func (b *StateBlock) GetPayload() []byte {
	if b.IsPrivate() {
		return b.GetPrivatePayload()
	}
	return b.Data
}

func (b *StateBlock) GetLink() Hash {
	return b.Link
}

func (b *StateBlock) GetSignature() Signature {
	return b.Signature
}

func (b *StateBlock) GetWork() Work {
	return b.Work
}

func (b *StateBlock) GetExtra() Hash {
	if b.Extra != nil {
		return *b.Extra
	}
	return ZeroHash
}

func (b *StateBlock) GetRepresentative() Address {
	return b.Representative
}

func (b *StateBlock) GetReceiver() []byte {
	return b.Receiver
}

func (b *StateBlock) GetSender() []byte {
	return b.Sender
}

func (b *StateBlock) GetMessage() Hash {
	if b.Message != nil {
		return *b.Message
	}
	return ZeroHash
}

func (b *StateBlock) GetTimestamp() int64 {
	return b.Timestamp
}

func (b *StateBlock) TotalBalance() Balance {
	balance := b.Balance
	balance = balance.Add(b.GetVote()).Add(b.GetNetwork()).Add(b.GetOracle()).Add(b.GetStorage())
	return balance
}

func (b *StateBlock) IsOpen() bool {
	return b.Previous.IsZero()
}

func (b *StateBlock) Root() Hash {
	if b.IsOpen() {
		return b.Address.ToHash()
	}
	return b.Previous
}

func (b *StateBlock) Parent() Hash {
	if b.IsOpen() {
		return b.Link
	}
	return b.Previous
}

func (b *StateBlock) Size() int {
	return b.Msgsize()
}

func (b *StateBlock) IsValid() bool {
	if b.IsOpen() {
		return b.Work.IsValid(Hash(b.Address))
	}
	return b.Work.IsValid(b.Previous)
}

func (b *StateBlock) Serialize() ([]byte, error) {
	return b.MarshalMsg(nil)
}

func (b *StateBlock) Deserialize(text []byte) error {
	_, err := b.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (b *StateBlock) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

func (b *StateBlock) IsReceiveBlock() bool {
	return b.Type == Receive || b.Type == Open || b.Type == ContractReward
}

func (b *StateBlock) IsSendBlock() bool {
	return b.Type == Send || b.Type == ContractSend
}

func (b *StateBlock) IsContractBlock() bool {
	return b.Type == ContractReward || b.Type == ContractSend || b.Type == ContractRefund || b.Type == ContractError
}

func (b *StateBlock) Clone() *StateBlock {
	clone := StateBlock{}
	blkBytes, _ := b.Serialize()
	_ = clone.Deserialize(blkBytes)
	clone.Flag = b.Flag

	// private tx fields
	clone.PrivateRecvRsp = b.PrivateRecvRsp
	if len(b.PrivatePayload) > 0 {
		clone.PrivatePayload = make([]byte, len(b.PrivatePayload))
		copy(clone.PrivatePayload, b.PrivatePayload)
	}

	return &clone
}

func (b *StateBlock) IsFromSync() bool {
	return b.Flag&BlockFlagSync == BlockFlagSync
}

func (b *StateBlock) SetFromSync() {
	b.Flag |= BlockFlagSync
}

func (b *StateBlock) IsPrivate() bool {
	if len(b.PrivateFrom) > 0 {
		return true
	}
	return false
}

func (b *StateBlock) IsRecipient() bool {
	if b.IsPrivate() {
		if len(b.PrivatePayload) == 0 {
			return false
		}
	}

	return true
}

func (b *StateBlock) GetPrivatePayload() []byte {
	return b.PrivatePayload
}

func (b *StateBlock) SetPrivatePayload(rawData []byte) {
	b.PrivatePayload = rawData
	b.PrivateRecvRsp = true
}

func (b *StateBlock) CheckPrivateRecvRsp() error {
	if b.IsPrivate() && !b.PrivateRecvRsp {
		return errors.New("private does not got recv rsp")
	}
	return nil
}

type StateBlockList []*StateBlock

func (bs *StateBlockList) Serialize() ([]byte, error) {
	return bs.MarshalMsg(nil)
}

func (bs *StateBlockList) Deserialize(text []byte) error {
	_, err := bs.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (b *StateBlock) ConvertToSchema() ([]Schema, error) {
	return []Schema{
		&BlockHash{
			Type:      b.Type.String(),
			Address:   b.Address.String(),
			Timestamp: b.Timestamp,
			Hash:      b.GetHash().String(),
		}}, nil
}

type BlockHash struct {
	Id        int64  `db:"id" typ:"integer"`
	Hash      string `db:"hash" typ:"char(64)"`
	Type      string `db:"type"  typ:"varchar(15)"`
	Address   string `db:"address" typ:"char(64)"`
	Timestamp int64  `db:"timestamp" typ:"integer"`
}

func (s *BlockHash) DeleteKey() string {
	return fmt.Sprintf("DELETE FROM BlockHash WHERE Hash = '%s'", s.Hash)
}

//
////go:generate msgp
//type BlockExtra struct {
//	KeyHash Hash    `msg:"key,extension" json:"key"`
//	Abi     []byte  `msg:"abi" json:"abi"`
//	Issuer  Address `msg:"issuer,extension" json:"issuer"`
//}
