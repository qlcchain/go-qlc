package types

import "fmt"

type AddressToken struct {
	Address Address
	Token   Hash
}

func (at AddressToken) String() string {
	return fmt.Sprintf("%s/%s", at.Address, at.Token)
}
