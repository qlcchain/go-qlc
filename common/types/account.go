/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"github.com/qlcchain/go-qlc/crypto/ed25519"
)

type Account struct {
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
}

//go:generate msgp
type TokenMeta struct {
	TokenAccount Address `msg:"tokenAccount,extension" json:"token_account"`
	Type         Hash    `msg:"type,extension" json:"type"`
	Header       Hash    `msg:"header,extension" json:"header"`
	RepBlock     Hash    `msg:"rep,extension" json:"rep"`
	OpenBlock    Hash    `msg:"open,extension" json:"open"`
	Balance      Balance `msg:"balance,extension" json:"balance"`
	Pending      Balance `msg:"pending,extension" json:"pending"`
	BelongTo     Address `msg:"account,extension" json:"account"`
	Modified     int64   `msg:"modified" json:"modified"`
	BlockCount   int64   `msg:"blockCount," json:"block_count"`
}

//go:generate msgp
type AccountMeta struct {
	Address Address      `msg:"account,extension" json:"account"`
	Tokens  []*TokenMeta `msg:"tokens" json:"tokens"`
}

// NewAccount creates a new account with the given private key.
func NewAccount(key ed25519.PrivateKey) *Account {
	return &Account{
		pubKey:  key.Public().(ed25519.PublicKey),
		privKey: key,
	}
}

// Address returns the public key of this account as an Address type.
func (a *Account) Address() Address {
	var address Address
	copy(address[:], a.pubKey)
	return address
}

func (a *Account) Sign(hash Hash) Signature {
	var sig Signature
	copy(sig[:], ed25519.Sign(a.privKey, hash[:]))
	return sig
}
