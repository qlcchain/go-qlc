/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"fmt"

	"github.com/qlcchain/go-qlc/crypto/ed25519"
)

type Account struct {
	pubKey  ed25519.PublicKey
	privKey ed25519.PrivateKey
}

//go:generate msgp
type TokenMeta struct {
	//TokenAccount Address `msg:"tokenAccount,extension" json:"token_account"`
	Type           Hash    `msg:"type,extension" json:"type"`
	Header         Hash    `msg:"header,extension" json:"header"`
	Representative Address `msg:"rep,extension" json:"representative"`
	OpenBlock      Hash    `msg:"open,extension" json:"open"`
	Balance        Balance `msg:"balance,extension" json:"balance"`
	BelongTo       Address `msg:"account,extension" json:"account"`
	Modified       int64   `msg:"modified" json:"modified"`
	BlockCount     int64   `msg:"blockCount," json:"blockCount"`
}

//go:generate msgp
type AccountMeta struct {
	Address     Address      `msg:"account,extension" json:"account"`
	CoinBalance Balance      `msg:"balance,extension" json:"balance"`
	CoinVote    Balance      `msg:"vote,extension" json:"vote"`
	CoinNetwork Balance      `msg:"network,extension" json:"network"`
	CoinStorage Balance      `msg:"storage,extension" json:"storage"`
	CoinOracle  Balance      `msg:"oracle,extension" json:"oracle"`
	Tokens      []*TokenMeta `msg:"tokens" json:"tokens"`
}

//Token get token meta by token type hash
func (am *AccountMeta) Token(tt Hash) *TokenMeta {
	for _, token := range am.Tokens {
		fmt.Printf("token: %p\n", token)
		if token.Type == tt {
			return token
		}
	}
	return nil
}

func (a *AccountMeta) VoteBalance() Balance {
	return a.CoinBalance.Add(a.CoinVote)
}

func (a *AccountMeta) TotalBalance() Balance {
	return a.CoinBalance.Add(a.CoinVote).Add(a.CoinStorage).Add(a.CoinNetwork).Add(a.CoinOracle)
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

func (a *Account) PrivateKey() ed25519.PrivateKey {
	return a.privKey
}

func (a *Account) Sign(hash Hash) Signature {
	var sig Signature
	copy(sig[:], ed25519.Sign(a.privKey, hash[:]))
	return sig
}

// String implements the fmt.Stringer interface.
func (a *Account) String() string {
	return fmt.Sprintf("Address: %s, Private key: %s", a.Address().String(), hex.EncodeToString(a.privKey))
}
