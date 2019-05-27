/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

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
	UncheckedKindLink
)

type SynchronizedKind byte

const (
	Synchronized SynchronizedKind = iota
	UnSynchronized
)

type UncheckedBlockWalkFunc func(block Block, link Hash, unCheckType UncheckedKind, sync SynchronizedKind) error
