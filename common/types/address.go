package types

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/qlcchain/go-qlc/common/util"
	"github.com/qlcchain/go-qlc/crypto/ed25519"
	"github.com/tinylib/msgp/msgp"
	"golang.org/x/crypto/blake2b"
)

func init() {
	msgp.RegisterExtension(AddressExtensionType, func() msgp.Extension { return new(Address) })
}

const (
	// AddressPrefix is the prefix of qlc addresses.
	AddressPrefix = "qlc_"
	// AddressSize represents the binary size of a qlc address (a public key).
	AddressSize         = ed25519.PublicKeySize
	addressChecksumSize = 5
	addressPrefixLen    = len(AddressPrefix)
	// AddressLen represents the string length of a qlc address.
	AddressLen = 60

	// The following 52 characters form the address, and the final
	// 8 are a checksum.
	hexAddressLength = addressPrefixLen + AddressLen
	// custom alphabet for base32 encoding
	addressEncodingAlphabet = "13456789abcdefghijkmnopqrstuwxyz"
)

var (
	// ZeroAddress
	ZeroAddress = Address{}

	MintageAddress, _    = HexToAddress("qlc_3qjky1ptg9qkzm8iertdzrnx9btjbaea33snh1w4g395xqqczye4kgcfyfs1")
	NEP5PledgeAddress, _ = HexToAddress("qlc_3fwi6r1fzjwmiys819pw8jxrcmcottsj4iq56kkgcmzi3b87596jwskwqrr5")
	RewardsAddress, _    = HexToAddress("qlc_3oinqggowa7f1rsjfmib476ggz6s4fp8578odjzerzztkrifqkqdz5zjztb3")

	ChainContractAddressList = []Address{NEP5PledgeAddress, MintageAddress, RewardsAddress}

	// AddressEncoding is a base32 encoding using addressEncodingAlphabet as its
	// alphabet.
	AddressEncoding = base32.NewEncoding(addressEncodingAlphabet)

	errAddressLen      = errors.New("bad address length")
	errAddressPrefix   = errors.New("bad address prefix")
	errAddressEncoding = errors.New("bad address encoding")
	errAddressChecksum = errors.New("bad address checksum")
)

//Address of account
//go:generate msgp
type Address [AddressSize]byte

//BytesToAddress convert byte array to Address
func BytesToAddress(b []byte) (Address, error) {
	var a Address
	err := a.SetBytes(b)
	return a, err
}

// HexToAddress convert hex address string to Address
func HexToAddress(hexStr string) (Address, error) {
	s := util.TrimQuotes(hexStr)
	if len(s) != hexAddressLength {
		return Address{}, errAddressLen
	}

	if !strings.HasPrefix(s, AddressPrefix) {
		return Address{}, errAddressPrefix
	}

	addr := s[addressPrefixLen:]

	key, err := AddressEncoding.DecodeString("1111" + addr[0:52])
	if err != nil {
		return Address{}, errAddressEncoding
	}

	checksum, err := AddressEncoding.DecodeString(addr[52:])
	if err != nil {
		return Address{}, errAddressEncoding
	}

	// strip off upper 24 bits (3 bytes). 20 padding was added by us,
	// 4 is unused as account is 256 bits.
	var address Address
	copy(address[:], key[3:])

	if !bytes.Equal(address.Checksum(), checksum) {
		return Address{}, errAddressChecksum
	}

	return address, nil
}

// IsValidHexAddress check Hex address string is valid
func IsValidHexAddress(hexStr string) bool {
	_, err := HexToAddress(hexStr)

	return err == nil
}

// PubToAddress  convert ed25519.PublicKey to Address
func PubToAddress(pub ed25519.PublicKey) Address {
	// Public key is 256bits, base32 must be multiple of 5 bits
	// to encode properly.
	addr, _ := BytesToAddress(pub)
	return addr
}

// GenerateAddress generate qlc address
func GenerateAddress() (Address, ed25519.PrivateKey, error) {
	pub, pri, err := ed25519.GenerateKey(rand.Reader)
	return PubToAddress(pub), pri, err
}

// KeypairFromPrivateKey generate key pair from private key
func KeypairFromPrivateKey(privateKey string) (ed25519.PublicKey, ed25519.PrivateKey) {
	privateBytes, _ := hex.DecodeString(privateKey)
	pub, priv, _ := ed25519.GenerateKey(bytes.NewReader(privateBytes))

	return pub, priv
}

// KeypairFromSeed generate key pair from seed
func KeypairFromSeed(seed string, index uint32) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	hash, err := blake2b.New(32, nil)
	if err != nil {
		return ed25519.PublicKey{}, ed25519.PrivateKey{}, errors.New("unable to create hash")
	}

	seedData, err := hex.DecodeString(seed)
	if err != nil {
		return ed25519.PublicKey{}, ed25519.PrivateKey{}, errors.New("invalid seed")
	}

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, index)

	hash.Write(seedData)
	hash.Write(bs)

	seedBytes := hash.Sum(nil)
	pub, priv, err := ed25519.GenerateKey(bytes.NewReader(seedBytes))

	if err != nil {
		return ed25519.PublicKey{}, ed25519.PrivateKey{}, errors.New("unable to generate ed25519 key")
	}

	return pub, priv, nil
}

// SetBytes new address bytes
func (addr *Address) SetBytes(b []byte) error {
	if length := len(b); length != AddressSize {
		return fmt.Errorf("error address size  %v", length)
	}
	copy(addr[:], b)
	return nil
}

//Bytes get Address byte array
func (addr Address) Bytes() []byte { return addr[:] }

// Checksum calculates the checksum for this address' public key.
func (addr Address) Checksum() []byte {
	hash, err := blake2b.New(addressChecksumSize, nil)
	if err != nil {
		panic(err)
	}

	hash.Write(addr[:])
	return util.ReverseBytes(hash.Sum(nil))
}

// IsZero check address is zero
func (addr Address) IsZero() bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}
	return true
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

//ExtensionType implements Extension.ExtensionType interface
func (addr *Address) ExtensionType() int8 {
	return AddressExtensionType
}

//ExtensionType implements Extension.Len interface
func (addr *Address) Len() int {
	return AddressSize
}

//ExtensionType implements Extension.MarshalBinaryTo interface
func (addr Address) MarshalBinaryTo(text []byte) error {
	copy(text, addr[:])
	return nil
}

//ExtensionType implements Extension.UnmarshalBinary interface
func (addr *Address) UnmarshalBinary(text []byte) error {
	size := len(text)
	if len(text) != AddressSize {
		return fmt.Errorf("bad address size: %d", size)
	}
	copy((*addr)[:], text)
	return nil
}

//UnmarshalText implements encoding.TextUnmarshaler
func (addr *Address) UnmarshalText(text []byte) error {
	tmp, err := HexToAddress(string(text))
	if err != nil {
		return err
	}
	copy((*addr)[:], tmp[:])
	return nil
}

//MarshalText implements encoding.Textmarshaler
func (addr Address) MarshalText() (text []byte, err error) {
	return []byte(addr.String()), nil
}

func (addr *Address) ToHash() Hash {
	return Hash(*addr)
}
