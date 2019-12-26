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
	InvalidSynchronized
)

func (s SynchronizedKind) String() string {
	switch s {
	case Synchronized:
		return "sync"
	case UnSynchronized:
		return "unsync"
	default:
		return ""
	}
}

func StringToSyncKind(str string) SynchronizedKind {
	switch str {
	case "sync":
		return Synchronized
	case "unsync":
		return UnSynchronized
	default:
		return InvalidSynchronized
	}
}

type UncheckedBlockWalkFunc func(block *StateBlock, link Hash, unCheckType UncheckedKind, sync SynchronizedKind) error
type GapPovBlockWalkFunc func(blocks StateBlockList, height uint64, sync SynchronizedKind) error
type GapPublishBlockWalkFunc func(block *StateBlock, sync SynchronizedKind) error
