package types

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/qlcchain/go-qlc/common"

	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"golang.org/x/crypto/blake2b"
	"strings"
)

const (
	// AddressPrefix is the prefix of Nano addresses.
	AddressPrefix = "qlc_"
	// AddressSize represents the binary size of a Nano address (a public key).
	AddressSize         = ed25519.PublicKeySize
	addressChecksumSize = 5
	addressPrefixLen    = len(AddressPrefix)
	// AddressLen represents the string length of a Nano address.
	AddressLen       = 2 * AddressSize
	hexAddressLength = addressPrefixLen + AddressLen
	// custom alphabet for base32 encoding
	AddressEncodingAlphabet = "13456789abcdefghijkmnopqrstuwxyz"
)

var (
	// AddressEncoding is a base32 encoding using NanoEncodingAlphabet as its
	// alphabet.
	AddressEncoding = base32.NewEncoding(AddressEncodingAlphabet)

	ErrAddressLen      = errors.New("bad address length")
	ErrAddressPrefix   = errors.New("bad address prefix")
	ErrAddressEncoding = errors.New("bad address encoding")
	ErrAddressChecksum = errors.New("bad address checksum")
)

type Address [AddressSize]byte

func BytesToAddress(b []byte) (Address, error) {
	var a Address
	err := a.SetBytes(b)
	return a, err
}

func HexToAddress(hexStr string) (Address, error) {
	if len(hexStr) != hexAddressLength {
		return Address{}, ErrAddressLen
	}

	if !strings.HasPrefix(hexStr, AddressPrefix) {
		return Address{}, ErrAddressPrefix
	}

	key, err := AddressEncoding.DecodeString("1111" + hexStr[addressPrefixLen:56])
	if err != nil {
		return Address{}, ErrAddressEncoding
	}

	checksum, err := AddressEncoding.DecodeString(hexStr[56:])
	if err != nil {
		return Address{}, ErrAddressEncoding
	}

	var address Address
	copy(address[:], key[3:])

	if !bytes.Equal(address.Checksum(), checksum) {
		return Address{}, ErrAddressChecksum
	}

	return address, nil
}

// Check Hex address string is valid
func IsValidHexAddress(hexStr string) bool {
	_, err := HexToAddress(hexStr)

	return err == nil
}

func PubkeyToAddress(pubkey []byte) Address {
	addr, _ := BytesToAddress(common.Hash(AddressSize, pubkey))
	return addr
}

func PrikeyToAddress(key ed25519.PrivateKey) Address {
	return PubkeyToAddress(key.PubByte())
}

func CreateAddress() (Address, ed25519.PrivateKey, error) {
	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	return PubkeyToAddress(pub), pri, err
}

func CreateAddressWithDeterministic(d [32]byte) (Address, ed25519.PrivateKey, error) {
	pub, pri, err := ed25519.GenerateKeyFromD(d)
	return PubkeyToAddress(pub), pri, err
}

// Set new address bytes
func (addr *Address) SetBytes(b []byte) error {
	if length := len(b); length != AddressSize {
		return fmt.Errorf("error address size  %v", length)
	}
	copy(addr[:], b)
	return nil
}

//Address byte array
func (addr Address) Bytes() []byte { return addr[:] }

// Checksum calculates the checksum for this address' public key.
func (addr Address) Checksum() []byte {
	hash, err := blake2b.New(addressChecksumSize, nil)
	if err != nil {
		panic(err)
	}

	hash.Write(addr[:])
	return common.ReverseBytes(hash.Sum(nil))
}

// String implements the fmt.Stringer interface.
func (addr Address) String() string {
	key := append([]byte{0, 0, 0}, addr[:]...)
	encodedKey := AddressEncoding.EncodeToString(key)[4:]
	encodedChecksum := AddressEncoding.EncodeToString(addr.Checksum())

	var buf bytes.Buffer
	buf.WriteString(AddressPrefix)
	buf.WriteString(encodedKey)
	buf.WriteString(encodedChecksum)
	return buf.String()
}

// Verify reports whether the given signature is valid for the given data.
func (addr Address) Verify(data []byte, signature []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(addr[:]), data, signature)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (addr Address) MarshalText() ([]byte, error) {
	return addr[:], nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (addr *Address) UnmarshalText(text []byte) error {
	a, err := HexToAddress(string(text))
	if err != nil {
		return err
	}

	*addr = a
	return nil
}
