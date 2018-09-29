package ledger

import (
	"github.com/qlcchain/go-qlc/common/types"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
)

type Account struct {
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
}

//go:generate msgp
type TokenMeta struct {
	Type      types.Hash    `msg:"type" json:"type"`
	Header    types.Hash    `msg:"header" json:"header"`
	RepBlock  types.Hash    `msg:"rep" json:"rep"`
	OpenBlock types.Hash    `msg:"open" json:"open"`
	Balance   types.Balance `msg:"balance" json:"balance"`
}

//go:generate msgp
type AccountMeta struct {
	Address types.Address `msg:"account" json:"account"`
	Tokens  []*TokenMeta  `msg:"tokens" json:"tokens"`
}

// NewAccount creates a new account with the given private key.
func NewAccount(key ed25519.PrivateKey) *Account {
	return &Account{
		pubKey:  key.Public().(ed25519.PublicKey),
		privKey: key,
	}
}

// Address returns the public key of this account as an Address type.
func (a *Account) Address() types.Address {
	var address types.Address
	copy(address[:], a.pubKey)
	return address
}

func (a *Account) Sign(hash types.Hash) types.Signature {
	var sig types.Signature
	copy(sig[:], ed25519.Sign(a.privKey, hash[:]))
	return sig
}
