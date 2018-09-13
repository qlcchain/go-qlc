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

	"golang.org/x/crypto/blake2b"
)

const (
	//WorkSize work size
	WorkSize = 8
)

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
func (w Work) IsValid(root Hash, threshold uint64) bool {
	worker, err := NewWorker(w, root, threshold)
	if err != nil {
		return false
	}
	return worker.IsValid()
}

// MarshalText implements the encoding.TextMarshaler interface.
func (w Work) MarshalText() ([]byte, error) {
	return []byte(w.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (w *Work) UnmarshalText(text []byte) error {
	// todo: fix this, this can't be right
	size := hex.DecodedLen(len(text))
	if size != WorkSize {
		return fmt.Errorf("bad work size: %d", size)
	}

	var work [WorkSize]byte
	if _, err := hex.Decode(work[:], text); err != nil {
		return err
	}

	*w = Work(binary.BigEndian.Uint64(work[:]))
	return nil
}

// String implements the fmt.Stringer interface.
func (w Work) String() string {
	var bytes [WorkSize]byte
	binary.BigEndian.PutUint64(bytes[:], uint64(w))
	return hex.EncodeToString(bytes[:])
}

// NewWorker create new worker
func NewWorker(work Work, root Hash, threshold uint64) (*Worker, error) {
	h, err := blake2b.New(WorkSize, nil)
	if err != nil {
		return &Worker{}, err
	}

	return &Worker{
		Threshold: threshold,
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

//Generate new work
func (w *Worker) Generate() Work {
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
