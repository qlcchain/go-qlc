/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types


const (
	PendingKeySize = AddressSize + HashSize
)

//go:generate msgp
type PendingKey struct {
	Address Address `msg:"account,extension" json:"account"`
	Hash    Hash    `msg:"hash,extension" json:"hash"`
}

//go:generate msgp
type PendingInfo struct {
	Source Address `msg:"source,extension" json:"source"`
	Amount Balance `msg:"amount,extension" json:"amount"`
	Type   Hash    `msg:"type,extension" json:"type"`
}

type UncheckedKind byte

const (
	UncheckedKindPrevious UncheckedKind = iota
	UncheckedKindSource
)

type UncheckedBlockWalkFunc func(block Block, kind UncheckedKind) error
