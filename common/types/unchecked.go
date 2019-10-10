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

func (pk *PendingKey) Serialize() ([]byte, error) {
	return pk.MarshalMsg(nil)
}

func (pk *PendingKey) Deserialize(text []byte) error {
	_, err := pk.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PendingInfo) Serialize() ([]byte, error) {
	return pi.MarshalMsg(nil)
}

func (pi *PendingInfo) Deserialize(text []byte) error {
	_, err := pi.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

type PendingKind byte

const (
	PendingNotUsed PendingKind = iota
	PendingUsed
)

type UncheckedKind byte

const (
	UncheckedKindPrevious UncheckedKind = iota
	UncheckedKindLink
	UncheckedKindTokenInfo
)

type SynchronizedKind byte

const (
	Synchronized SynchronizedKind = iota
	UnSynchronized
)

func (s SynchronizedKind) ToString() string {
	switch s {
	case Synchronized:
		return "sync"
	case UnSynchronized:
		return "unsync"
	default:
		return ""
	}
}

type UncheckedBlockWalkFunc func(block *StateBlock, link Hash, unCheckType UncheckedKind, sync SynchronizedKind) error
