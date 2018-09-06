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
	"golang.org/x/crypto/blake2b"
	"hash"
)

const (
	WorkSize = 8
)

type Work uint64

type Worker struct {
	Threshold uint64
	root      *Hash
	work      Work
	hash      hash.Hash
}

func (w Work) Valid(root Hash, threshold uint64) bool {
	return NewWorker(w, root, threshold).Valid()
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

func NewWorker(work Work, root Hash, threshold uint64) *Worker {
	hash, err := blake2b.New(WorkSize, nil)
	if err != nil {
		panic(err)
	}

	return &Worker{
		Threshold: threshold,
		root:      &root,
		work:      work,
		hash:      hash,
	}
}

func (w *Worker) Valid() bool {
	var workBytes [WorkSize]byte
	binary.LittleEndian.PutUint64(workBytes[:], uint64(w.work))

	w.hash.Reset()
	w.hash.Write(workBytes[:])
	w.hash.Write(w.root[:])

	sum := w.hash.Sum(nil)
	value := binary.LittleEndian.Uint64(sum)
	return value >= w.Threshold
}

func (w *Worker) Generate() Work {
	for {
		if w.Valid() {
			return w.work
		}
		w.work++
	}
}

func (w *Worker) Reset() {
	w.work = 0
	w.hash.Reset()
}
