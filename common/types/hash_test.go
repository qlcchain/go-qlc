/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/qlcchain/go-qlc/crypto/random"
)

func TestHash(t *testing.T) {
	var hash Hash
	s := "2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"
	err := hash.Of(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	upper := strings.ToUpper(hash.String())
	if upper != s {
		t.Errorf("expect:%s but %s", s, upper)
	}
}

func TestHash_MarshalJSON(t *testing.T) {
	var hash Hash
	s := `"2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"`
	err := hash.Of(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	b, err := json.Marshal(&hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}

func TestHash_UnmarshalJSON(t *testing.T) {
	var hash Hash
	s := `"2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"`

	err := json.Unmarshal([]byte(s), &hash)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(hash)
}

func TestNewHash(t *testing.T) {
	hash, err := NewHash("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hash)
}

func TestHash_String(t *testing.T) {
	h := Hash{}
	t.Log(h.String())
	if !h.IsZero() {
		t.Fatal("zero hash error")
	}
}

func TestBytesToHash(t *testing.T) {
	bytes, err := hex.DecodeString("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BytesToHash(bytes); err != nil {
		t.Fatal(err)
	}
	bytes = append(bytes, 0x01)
	if _, err := BytesToHash(bytes); err == nil {
		t.Fatal("bytes2hash failed")
	}
}

func TestHash_Serialize(t *testing.T) {
	var hash Hash
	s := `"2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"`
	err := hash.Of(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Log(hash)

	v, err := hash.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(v)
	var h Hash
	if err := h.Deserialize(v); err != nil {
		t.Fatal(err)
	}
	t.Log(h)
	if hash != h {
		t.Fatal()
	}
}

func TestHash_Bytes(t *testing.T) {
	var hash Hash
	s := `"2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E"`
	err := hash.Of(s)
	if err != nil {
		t.Errorf("%v", err)
	}

	b2, _ := hex.DecodeString("2C353DA641277FD8379354307A54BECE090C51E52FB460EA5A8674B702BDCE5E")

	if !bytes.Equal(hash.Bytes(), b2) {
		t.Fatalf("%v,%v", hash[:], b2)
	}
}

func TestHash_Cmp(t *testing.T) {
	d1 := make([]byte, HashSize)
	d1[0] = 0x01
	h1, _ := BytesToHash(d1)
	d2 := make([]byte, HashSize)
	d2[0] = 0x02
	h2, _ := BytesToHash(d2)
	if cmp := h1.Cmp(h2); cmp != -1 {
		t.Fatalf("%d", cmp)
	}
}

func TestHash_ReverseByte(t *testing.T) {
	d1 := make([]byte, HashSize)
	d1[0] = 0x01
	h1, _ := BytesToHash(d1)
	h2 := h1.ReverseByte()
	if h2[0] != 0x0 && h2[31] != 0x01 {
		t.Fatalf("%v", h2)
	}
}

func TestHashData(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)
	if h := HashData(d1); h.IsZero() {
		t.Fatal()
	}
}

func TestHashBytes(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)
	d2 := make([]byte, 10)
	_ = random.Bytes(d2)

	if hash, err := HashBytes(d1, d2); err != nil {
		t.Fatal(err)
	} else {
		if hash.IsZero() {
			t.Fatal()
		}
	}
}

func TestSha256DHashData(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)

	if h := Sha256DHashData(d1); h.IsZero() {
		t.Fatal()
	}
}

func TestScryptHashData(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)

	if h := ScryptHashData(d1); h.IsZero() {
		t.Fatal()
	}
}

func TestSha256DHashBytes(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)
	d2 := make([]byte, 10)
	_ = random.Bytes(d2)

	if h, err := Sha256DHashBytes(d1, d2); err != nil {
		t.Fatal(err)
	} else if h.IsZero() {
		t.Fatal()
	}
}

func TestHash_ToBigInt(t *testing.T) {
	d1 := make([]byte, HashSize)
	_ = random.Bytes(d1)

	if hash, err := BytesToHash(d1); err != nil {
		t.Fatal(err)
	} else {
		i := hash.ToBigInt()
		h2 := &Hash{}
		if err := h2.FromBigInt(i); err != nil {
			t.Fatal(err)
		} else {
			if !reflect.DeepEqual(&hash, h2) {
				t.Fatalf("exp: %v, act: %v", &hash, h2)
			}
		}
	}
}

func TestScryptHashBytes(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)
	d2 := make([]byte, 10)
	_ = random.Bytes(d2)

	if h, err := ScryptHashBytes(d1, d2); err != nil {
		t.Fatal(err)
	} else if h.IsZero() {
		t.Fatal()
	}
}

func TestX11HashData(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)

	if h := X11HashData(d1); h.IsZero() {
		t.Fatal()
	}
}

func TestSha256HashData(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)

	if h, err := Sha256HashData(d1); err != nil {
		t.Fatal()
	} else if h.IsZero() {
		t.Fatal()
	}
}

func TestHybridHashData(t *testing.T) {
	d1 := make([]byte, 30)
	_ = random.Bytes(d1)

	if h := HybridHashData(d1); h.IsZero() {
		t.Fatal()
	}
}
