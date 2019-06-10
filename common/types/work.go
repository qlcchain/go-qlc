/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"strconv"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

func init() {
	msgp.RegisterExtension(WorkExtensionType, func() msgp.Extension { return new(Work) })
}

const (
	//WorkSize work size
	WorkSize = 8
)

var WorkThreshold = uint64(0)

// Work PoW work
type Work uint64

// Worker PoW
type Worker struct {
	Threshold uint64
	root      *Hash
	work      Work
	hash      hash.Hash
}

// IsValid check work is valid
func (w Work) IsValid(root Hash) bool {
	worker, err := NewWorker(w, root)
	if err != nil {
		return false
	}
	return worker.IsValid()
}

// String implements the fmt.Stringer interface.
func (w Work) String() string {
	var bytes [WorkSize]byte
	binary.BigEndian.PutUint64(bytes[:], uint64(w))
	return hex.EncodeToString(bytes[:])
}

//ParseWorkHexString create Work from hex string
func (w *Work) ParseWorkHexString(hexString string) error {
	s := util.TrimQuotes(hexString)
	work, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return err
	}
	*w = Work(work)
	return nil
}

//ExtensionType implements Extension.ExtensionType interface
func (w *Work) ExtensionType() int8 {
	return WorkExtensionType
}

//ExtensionType implements Extension.Len interface
func (w *Work) Len() int {
	return WorkSize
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (w *Work) MarshalBinaryTo(text []byte) error {
	var bytes [WorkSize]byte
	binary.BigEndian.PutUint64(bytes[:], uint64(*w))
	copy(text, bytes[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (w *Work) UnmarshalBinary(text []byte) error {
	size := len(text)
	if size != WorkSize {
		return fmt.Errorf("bad work size: %d", size)
	}

	*w = Work(binary.BigEndian.Uint64(text[:]))
	return nil
}

func (w *Work) UnmarshalText(text []byte) error {
	return w.ParseWorkHexString(string(text))
}

// MarshalText implements the encoding.TextMarshaler interface.
func (w Work) MarshalText() (text []byte, err error) {
	return []byte(w.String()), nil
}

//MarshalJSON implements json.Marshaler interface
//func (w *Work) MarshalJSON() ([]byte, error) {
//	return []byte(w.String()), nil
//}
//
//// UnmarshalJSON implements json.UnMarshaler interface
//func (w *Work) UnmarshalJSON(b []byte) error {
//	return w.ParseWorkHexString(string(b))
//}

// NewWorker create new worker
func NewWorker(work Work, root Hash) (*Worker, error) {
	h, err := blake2b.New(WorkSize, nil)
	if err != nil {
		return &Worker{}, err
	}

	return &Worker{
		Threshold: WorkThreshold,
		root:      &root,
		work:      work,
		hash:      h,
	}, nil
}

//IsValid check work is valid
func (w *Worker) IsValid() bool {
	var workBytes [WorkSize]byte
	binary.LittleEndian.PutUint64(workBytes[:], uint64(w.work))

	w.hash.Reset()
	w.hash.Write(workBytes[:])
	w.hash.Write(w.root[:])

	sum := w.hash.Sum(nil)
	value := binary.LittleEndian.Uint64(sum)
	return value >= w.Threshold
}

//NewWork generate new work
func (w *Worker) NewWork() Work {
	for {
		if w.IsValid() {
			return w.work
		}
		w.work++
	}
}

//Reset worker
func (w *Worker) Reset() {
	w.work = 0
	w.hash.Reset()
}
