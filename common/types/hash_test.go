/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
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
