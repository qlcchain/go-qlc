/*
 * Copyright (c) 2019 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package types

import "fmt"

type ContractKey struct {
	ContractAddress Address `json:"contractAddress"`
	AccountAddress  Address `json:"accountAddress"`
	Hash            Hash    `json:"hash,omitempty"`
	Suffix          []byte  `json:"suffix,omitempty"`
}

const (
	hashType byte = iota
	suffixType
)

const minLen = AddressSize * 2

func (z ContractKey) Serialize() ([]byte, error) {
	var data []byte
	data = append(data, z.ContractAddress[:]...)
	data = append(data, z.AccountAddress[:]...)
	if !z.Hash.IsZero() {
		data = append(data, hashType)
		data = append(data, z.Hash[:]...)
	}
	if len(z.Suffix) > 0 {
		data = append(data, suffixType)
		data = append(data, z.Suffix[:]...)
	}
	return data, nil
}

func (z *ContractKey) Deserialize(text []byte) (err error) {
	z.Suffix = nil
	l := len(text)
	if l < minLen {
		return fmt.Errorf("invalid len %d", l)
	}
	idx := AddressSize
	if z.ContractAddress, err = BytesToAddress(text[:idx]); err != nil {
		return err
	}
	if z.AccountAddress, err = BytesToAddress(text[idx : idx+AddressSize]); err != nil {
		return err
	}
	idx += AddressSize

	if idx < l && text[idx] == hashType {
		idx += 1
		if h, err := BytesToHash(text[idx : idx+HashSize]); err != nil {
			return err
		} else {
			z.Hash = h
		}
		idx += HashSize
	}
	if idx < l && text[idx] == suffixType {
		idx += 1
		z.Suffix = text[idx:]
	}
	return nil
}

func (z *ContractKey) String() string {
	return fmt.Sprintf("ca:%s, a:%s, h:%s, s: %v", z.ContractAddress, z.AccountAddress, z.Hash, z.Suffix)
}
