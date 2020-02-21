/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import "encoding/json"

type UncheckedKind byte

const (
	UncheckedKindPrevious UncheckedKind = iota
	UncheckedKindLink
	UncheckedKindTokenInfo
	UncheckedKindPublish
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
type GapPovBlockWalkFunc func(block *StateBlock, height uint64, sync SynchronizedKind) error
type GapPublishBlockWalkFunc func(block *StateBlock, sync SynchronizedKind) error

type Unchecked struct {
	Block *StateBlock      `msg:"block" json:"block"`
	Kind  SynchronizedKind `msg:"kind" json:"kind"`
}

func (u *Unchecked) Serialize() ([]byte, error) {
	return u.MarshalMsg(nil)
}

func (u *Unchecked) Deserialize(text []byte) error {
	_, err := u.UnmarshalMsg(text)
	if err != nil {
		return err
	}
	return nil
}

func (u *Unchecked) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}
