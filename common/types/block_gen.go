package types

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *BlockExtra) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "key":
			err = dc.ReadExtension(&z.KeyHash)
			if err != nil {
				return
			}
		case "abi":
			z.Abi, err = dc.ReadBytes(z.Abi)
			if err != nil {
				return
			}
		case "issuer":
			err = dc.ReadExtension(&z.Issuer)
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
func (z *BlockExtra) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "key"
	err = en.Append(0x83, 0xa3, 0x6b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.KeyHash)
	if err != nil {
		return
	}
	// write "abi"
	err = en.Append(0xa3, 0x61, 0x62, 0x69)
	if err != nil {
		return
	}
	err = en.WriteBytes(z.Abi)
	if err != nil {
		return
	}
	// write "issuer"
	err = en.Append(0xa6, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Issuer)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *BlockExtra) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "key"
	o = append(o, 0x83, 0xa3, 0x6b, 0x65, 0x79)
	o, err = msgp.AppendExtension(o, &z.KeyHash)
	if err != nil {
		return
	}
	// string "abi"
	o = append(o, 0xa3, 0x61, 0x62, 0x69)
	o = msgp.AppendBytes(o, z.Abi)
	// string "issuer"
	o = append(o, 0xa6, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72)
	o, err = msgp.AppendExtension(o, &z.Issuer)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BlockExtra) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "key":
			bts, err = msgp.ReadExtensionBytes(bts, &z.KeyHash)
			if err != nil {
				return
			}
		case "abi":
			z.Abi, bts, err = msgp.ReadBytesBytes(bts, z.Abi)
			if err != nil {
				return
			}
		case "issuer":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Issuer)
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
func (z *BlockExtra) Msgsize() (s int) {
	s = 1 + 4 + msgp.ExtensionPrefixSize + z.KeyHash.Len() + 4 + msgp.BytesPrefixSize + len(z.Abi) + 7 + msgp.ExtensionPrefixSize + z.Issuer.Len()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *BlockType) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 string
		zb0001, err = dc.ReadString()
		if err != nil {
			return
		}
		(*z) = parseString(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z BlockType) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteString((BlockType).String(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z BlockType) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendString(o, (BlockType).String(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *BlockType) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 string
		zb0001, bts, err = msgp.ReadStringBytes(bts)
		if err != nil {
			return
		}
		(*z) = parseString(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z BlockType) Msgsize() (s int) {
	s = msgp.StringPrefixSize + len((BlockType).String(z))
	return
}

// DecodeMsg implements msgp.Decodable
func (z *CommonBlock) DecodeMsg(dc *msgp.Reader) (err error) {
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
			{
				var zb0002 string
				zb0002, err = dc.ReadString()
				if err != nil {
					return
				}
				z.Type = parseString(zb0002)
			}
		case "address":
			err = dc.ReadExtension(&z.Address)
			if err != nil {
				return
			}
		case "previous":
			err = dc.ReadExtension(&z.Previous)
			if err != nil {
				return
			}
		case "extra":
			err = dc.ReadExtension(&z.Extra)
			if err != nil {
				return
			}
		case "work":
			err = dc.ReadExtension(&z.Work)
			if err != nil {
				return
			}
		case "signature":
			err = dc.ReadExtension(&z.Signature)
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
func (z *CommonBlock) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 6
	// write "type"
	err = en.Append(0x86, 0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString((BlockType).String(z.Type))
	if err != nil {
		return
	}
	// write "address"
	err = en.Append(0xa7, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Address)
	if err != nil {
		return
	}
	// write "previous"
	err = en.Append(0xa8, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Previous)
	if err != nil {
		return
	}
	// write "extra"
	err = en.Append(0xa5, 0x65, 0x78, 0x74, 0x72, 0x61)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Extra)
	if err != nil {
		return
	}
	// write "work"
	err = en.Append(0xa4, 0x77, 0x6f, 0x72, 0x6b)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Work)
	if err != nil {
		return
	}
	// write "signature"
	err = en.Append(0xa9, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Signature)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *CommonBlock) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 6
	// string "type"
	o = append(o, 0x86, 0xa4, 0x74, 0x79, 0x70, 0x65)
	o = msgp.AppendString(o, (BlockType).String(z.Type))
	// string "address"
	o = append(o, 0xa7, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73)
	o, err = msgp.AppendExtension(o, &z.Address)
	if err != nil {
		return
	}
	// string "previous"
	o = append(o, 0xa8, 0x70, 0x72, 0x65, 0x76, 0x69, 0x6f, 0x75, 0x73)
	o, err = msgp.AppendExtension(o, &z.Previous)
	if err != nil {
		return
	}
	// string "extra"
	o = append(o, 0xa5, 0x65, 0x78, 0x74, 0x72, 0x61)
	o, err = msgp.AppendExtension(o, &z.Extra)
	if err != nil {
		return
	}
	// string "work"
	o = append(o, 0xa4, 0x77, 0x6f, 0x72, 0x6b)
	o, err = msgp.AppendExtension(o, &z.Work)
	if err != nil {
		return
	}
	// string "signature"
	o = append(o, 0xa9, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65)
	o, err = msgp.AppendExtension(o, &z.Signature)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *CommonBlock) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
			{
				var zb0002 string
				zb0002, bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
				z.Type = parseString(zb0002)
			}
		case "address":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Address)
			if err != nil {
				return
			}
		case "previous":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Previous)
			if err != nil {
				return
			}
		case "extra":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Extra)
			if err != nil {
				return
			}
		case "work":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Work)
			if err != nil {
				return
			}
		case "signature":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Signature)
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
func (z *CommonBlock) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len((BlockType).String(z.Type)) + 8 + msgp.ExtensionPrefixSize + z.Address.Len() + 9 + msgp.ExtensionPrefixSize + z.Previous.Len() + 6 + msgp.ExtensionPrefixSize + z.Extra.Len() + 5 + msgp.ExtensionPrefixSize + z.Work.Len() + 10 + msgp.ExtensionPrefixSize + z.Signature.Len()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SmartContractBlock) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "internalAccount":
			err = dc.ReadExtension(&z.InternalAccount)
			if err != nil {
				return
			}
		case "contract":
			err = dc.ReadExtension(&z.Abi)
			if err != nil {
				return
			}
		case "issuer":
			err = dc.ReadExtension(&z.Issuer)
			if err != nil {
				return
			}
		case "CommonBlock":
			err = z.CommonBlock.DecodeMsg(dc)
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
func (z *SmartContractBlock) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 4
	// write "internalAccount"
	err = en.Append(0x84, 0xaf, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.InternalAccount)
	if err != nil {
		return
	}
	// write "contract"
	err = en.Append(0xa8, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Abi)
	if err != nil {
		return
	}
	// write "issuer"
	err = en.Append(0xa6, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Issuer)
	if err != nil {
		return
	}
	// write "CommonBlock"
	err = en.Append(0xab, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	if err != nil {
		return
	}
	err = z.CommonBlock.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *SmartContractBlock) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 4
	// string "internalAccount"
	o = append(o, 0x84, 0xaf, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74)
	o, err = msgp.AppendExtension(o, &z.InternalAccount)
	if err != nil {
		return
	}
	// string "contract"
	o = append(o, 0xa8, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74)
	o, err = msgp.AppendExtension(o, &z.Abi)
	if err != nil {
		return
	}
	// string "issuer"
	o = append(o, 0xa6, 0x69, 0x73, 0x73, 0x75, 0x65, 0x72)
	o, err = msgp.AppendExtension(o, &z.Issuer)
	if err != nil {
		return
	}
	// string "CommonBlock"
	o = append(o, 0xab, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	o, err = z.CommonBlock.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SmartContractBlock) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "internalAccount":
			bts, err = msgp.ReadExtensionBytes(bts, &z.InternalAccount)
			if err != nil {
				return
			}
		case "contract":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Abi)
			if err != nil {
				return
			}
		case "issuer":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Issuer)
			if err != nil {
				return
			}
		case "CommonBlock":
			bts, err = z.CommonBlock.UnmarshalMsg(bts)
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
func (z *SmartContractBlock) Msgsize() (s int) {
	s = 1 + 16 + msgp.ExtensionPrefixSize + z.InternalAccount.Len() + 9 + msgp.ExtensionPrefixSize + z.Abi.Len() + 7 + msgp.ExtensionPrefixSize + z.Issuer.Len() + 12 + z.CommonBlock.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *StateBlock) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "token":
			err = dc.ReadExtension(&z.Token)
			if err != nil {
				return
			}
		case "balance":
			err = dc.ReadExtension(&z.Balance)
			if err != nil {
				return
			}
		case "link":
			err = dc.ReadExtension(&z.Link)
			if err != nil {
				return
			}
		case "representative":
			err = dc.ReadExtension(&z.Representative)
			if err != nil {
				return
			}
		case "CommonBlock":
			err = z.CommonBlock.DecodeMsg(dc)
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
func (z *StateBlock) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "token"
	err = en.Append(0x85, 0xa5, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Token)
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
	// write "link"
	err = en.Append(0xa4, 0x6c, 0x69, 0x6e, 0x6b)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Link)
	if err != nil {
		return
	}
	// write "representative"
	err = en.Append(0xae, 0x72, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x76, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Representative)
	if err != nil {
		return
	}
	// write "CommonBlock"
	err = en.Append(0xab, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	if err != nil {
		return
	}
	err = z.CommonBlock.EncodeMsg(en)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *StateBlock) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "token"
	o = append(o, 0x85, 0xa5, 0x74, 0x6f, 0x6b, 0x65, 0x6e)
	o, err = msgp.AppendExtension(o, &z.Token)
	if err != nil {
		return
	}
	// string "balance"
	o = append(o, 0xa7, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o, err = msgp.AppendExtension(o, &z.Balance)
	if err != nil {
		return
	}
	// string "link"
	o = append(o, 0xa4, 0x6c, 0x69, 0x6e, 0x6b)
	o, err = msgp.AppendExtension(o, &z.Link)
	if err != nil {
		return
	}
	// string "representative"
	o = append(o, 0xae, 0x72, 0x65, 0x70, 0x72, 0x65, 0x73, 0x65, 0x6e, 0x74, 0x61, 0x74, 0x69, 0x76, 0x65)
	o, err = msgp.AppendExtension(o, &z.Representative)
	if err != nil {
		return
	}
	// string "CommonBlock"
	o = append(o, 0xab, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x42, 0x6c, 0x6f, 0x63, 0x6b)
	o, err = z.CommonBlock.MarshalMsg(o)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *StateBlock) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "token":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Token)
			if err != nil {
				return
			}
		case "balance":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Balance)
			if err != nil {
				return
			}
		case "link":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Link)
			if err != nil {
				return
			}
		case "representative":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Representative)
			if err != nil {
				return
			}
		case "CommonBlock":
			bts, err = z.CommonBlock.UnmarshalMsg(bts)
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
func (z *StateBlock) Msgsize() (s int) {
	s = 1 + 6 + msgp.ExtensionPrefixSize + z.Token.Len() + 8 + msgp.ExtensionPrefixSize + z.Balance.Len() + 5 + msgp.ExtensionPrefixSize + z.Link.Len() + 15 + msgp.ExtensionPrefixSize + z.Representative.Len() + 12 + z.CommonBlock.Msgsize()
	return
}
