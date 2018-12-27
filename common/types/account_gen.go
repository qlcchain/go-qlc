package types

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Account) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z Account) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 0
	err = en.Append(0x80)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Account) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 0
	o = append(o, 0x80)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Account) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Account) Msgsize() (s int) {
	s = 1
	return
}

// DecodeMsg implements msgp.Decodable
func (z *AccountMeta) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "account":
			err = dc.ReadExtension(&z.Address)
			if err != nil {
				return
			}
		case "tokens":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.Tokens) >= int(zb0002) {
				z.Tokens = (z.Tokens)[:zb0002]
			} else {
				z.Tokens = make([]*TokenMeta, zb0002)
			}
			for za0001 := range z.Tokens {
				if dc.IsNil() {
					err = dc.ReadNil()
					if err != nil {
						return
					}
					z.Tokens[za0001] = nil
				} else {
					if z.Tokens[za0001] == nil {
						z.Tokens[za0001] = new(TokenMeta)
					}
					err = z.Tokens[za0001].DecodeMsg(dc)
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *AccountMeta) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 2
	// write "account"
	err = en.Append(0x82, 0xa7, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Address)
	if err != nil {
		return
	}
	// write "tokens"
	err = en.Append(0xa6, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73)
	if err != nil {
		return
	}
	err = en.WriteArrayHeader(uint32(len(z.Tokens)))
	if err != nil {
		return
	}
	for za0001 := range z.Tokens {
		if z.Tokens[za0001] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z.Tokens[za0001].EncodeMsg(en)
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *AccountMeta) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "account"
	o = append(o, 0x82, 0xa7, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o, err = msgp.AppendExtension(o, &z.Address)
	if err != nil {
		return
	}
	// string "tokens"
	o = append(o, 0xa6, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.Tokens)))
	for za0001 := range z.Tokens {
		if z.Tokens[za0001] == nil {
			o = msgp.AppendNil(o)
		} else {
			o, err = z.Tokens[za0001].MarshalMsg(o)
			if err != nil {
				return
			}
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *AccountMeta) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "account":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Address)
			if err != nil {
				return
			}
		case "tokens":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.Tokens) >= int(zb0002) {
				z.Tokens = (z.Tokens)[:zb0002]
			} else {
				z.Tokens = make([]*TokenMeta, zb0002)
			}
			for za0001 := range z.Tokens {
				if msgp.IsNil(bts) {
					bts, err = msgp.ReadNilBytes(bts)
					if err != nil {
						return
					}
					z.Tokens[za0001] = nil
				} else {
					if z.Tokens[za0001] == nil {
						z.Tokens[za0001] = new(TokenMeta)
					}
					bts, err = z.Tokens[za0001].UnmarshalMsg(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *AccountMeta) Msgsize() (s int) {
	s = 1 + 8 + msgp.ExtensionPrefixSize + z.Address.Len() + 7 + msgp.ArrayHeaderSize
	for za0001 := range z.Tokens {
		if z.Tokens[za0001] == nil {
			s += msgp.NilSize
		} else {
			s += z.Tokens[za0001].Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *TokenMeta) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "type":
			err = dc.ReadExtension(&z.Type)
			if err != nil {
				return
			}
		case "header":
			err = dc.ReadExtension(&z.Header)
			if err != nil {
				return
			}
		case "rep":
			err = dc.ReadExtension(&z.Representative)
			if err != nil {
				return
			}
		case "open":
			err = dc.ReadExtension(&z.OpenBlock)
			if err != nil {
				return
			}
		case "balance":
			err = dc.ReadExtension(&z.Balance)
			if err != nil {
				return
			}
		case "account":
			err = dc.ReadExtension(&z.BelongTo)
			if err != nil {
				return
			}
		case "modified":
			z.Modified, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "blockCount":
			z.BlockCount, err = dc.ReadInt64()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *TokenMeta) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 8
	// write "type"
	err = en.Append(0x88, 0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Type)
	if err != nil {
		return
	}
	// write "header"
	err = en.Append(0xa6, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Header)
	if err != nil {
		return
	}
	// write "rep"
	err = en.Append(0xa3, 0x72, 0x65, 0x70)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Representative)
	if err != nil {
		return
	}
	// write "open"
	err = en.Append(0xa4, 0x6f, 0x70, 0x65, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.OpenBlock)
	if err != nil {
		return
	}
	// write "balance"
	err = en.Append(0xa7, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Balance)
	if err != nil {
		return
	}
	// write "account"
	err = en.Append(0xa7, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.BelongTo)
	if err != nil {
		return
	}
	// write "modified"
	err = en.Append(0xa8, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Modified)
	if err != nil {
		return
	}
	// write "blockCount"
	err = en.Append(0xaa, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.BlockCount)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *TokenMeta) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 8
	// string "type"
	o = append(o, 0x88, 0xa4, 0x74, 0x79, 0x70, 0x65)
	o, err = msgp.AppendExtension(o, &z.Type)
	if err != nil {
		return
	}
	// string "header"
	o = append(o, 0xa6, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72)
	o, err = msgp.AppendExtension(o, &z.Header)
	if err != nil {
		return
	}
	// string "rep"
	o = append(o, 0xa3, 0x72, 0x65, 0x70)
	o, err = msgp.AppendExtension(o, &z.Representative)
	if err != nil {
		return
	}
	// string "open"
	o = append(o, 0xa4, 0x6f, 0x70, 0x65, 0x6e)
	o, err = msgp.AppendExtension(o, &z.OpenBlock)
	if err != nil {
		return
	}
	// string "balance"
	o = append(o, 0xa7, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o, err = msgp.AppendExtension(o, &z.Balance)
	if err != nil {
		return
	}
	// string "account"
	o = append(o, 0xa7, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o, err = msgp.AppendExtension(o, &z.BelongTo)
	if err != nil {
		return
	}
	// string "modified"
	o = append(o, 0xa8, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64)
	o = msgp.AppendInt64(o, z.Modified)
	// string "blockCount"
	o = append(o, 0xaa, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x43, 0x6f, 0x75, 0x6e, 0x74)
	o = msgp.AppendInt64(o, z.BlockCount)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *TokenMeta) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "type":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Type)
			if err != nil {
				return
			}
		case "header":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Header)
			if err != nil {
				return
			}
		case "rep":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Representative)
			if err != nil {
				return
			}
		case "open":
			bts, err = msgp.ReadExtensionBytes(bts, &z.OpenBlock)
			if err != nil {
				return
			}
		case "balance":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Balance)
			if err != nil {
				return
			}
		case "account":
			bts, err = msgp.ReadExtensionBytes(bts, &z.BelongTo)
			if err != nil {
				return
			}
		case "modified":
			z.Modified, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "blockCount":
			z.BlockCount, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *TokenMeta) Msgsize() (s int) {
	s = 1 + 5 + msgp.ExtensionPrefixSize + z.Type.Len() + 7 + msgp.ExtensionPrefixSize + z.Header.Len() + 4 + msgp.ExtensionPrefixSize + z.Representative.Len() + 5 + msgp.ExtensionPrefixSize + z.OpenBlock.Len() + 8 + msgp.ExtensionPrefixSize + z.Balance.Len() + 8 + msgp.ExtensionPrefixSize + z.BelongTo.Len() + 9 + msgp.Int64Size + 11 + msgp.Int64Size
	return
}
