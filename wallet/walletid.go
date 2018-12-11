/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package wallet

import (
	"github.com/google/uuid"
)

type WalletId struct {
	id uuid.UUID
}

func NewWalletId() WalletId {
	return WalletId{id: uuid.New()}
}

func String2WalletId(s string) (WalletId, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return WalletId{}, err
	}
	return WalletId{id: id}, nil
}

//
//func Bytes2WalletId(b []byte) (WalletId, error) {
//	id, err := uuid.ParseBytes(b)
//	if err != nil {
//		return WalletId{}, err
//	}
//	return WalletId{id: id}, nil
//}

func (w WalletId) String() string {
	return w.id.String()
}

// MarshalText implements encoding.TextMarshaler.
func (w WalletId) MarshalText() ([]byte, error) {
	return w.id.MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (w *WalletId) UnmarshalText(data []byte) error {
	return w.id.UnmarshalText(data)
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (w WalletId) MarshalBinary() ([]byte, error) {
	return w.id.MarshalBinary()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (w *WalletId) UnmarshalBinary(data []byte) error {
	return w.id.UnmarshalBinary(data)
}
