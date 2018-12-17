/*
 * Copyright (c) 2018 QLC Chain Team
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package crypto

import "github.com/awnumar/memguard"

type SecureString struct {
	buff *memguard.LockedBuffer
	size int
}

//NewSecureString from string
func NewSecureString(passphrase string) (*SecureString, error) {
	bytes := []byte(passphrase)
	buffer, err := memguard.NewImmutableFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	return &SecureString{
		buff: buffer,
		size: len(bytes),
	}, nil
}

//Bytes get bytes from string
func (ss *SecureString) Bytes() []byte {
	return ss.buff.Buffer()
}

//Destroy string buffer
func (ss *SecureString) Destroy() {
	ss.buff.Destroy()
}
