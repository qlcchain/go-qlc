package types

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Benefit) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "balance":
			err = dc.ReadExtension(&z.Balance)
			if err != nil {
				return
			}
		case "vote":
			err = dc.ReadExtension(&z.Vote)
			if err != nil {
				return
			}
		case "network":
			err = dc.ReadExtension(&z.Network)
			if err != nil {
				return
			}
		case "storage":
			err = dc.ReadExtension(&z.Storage)
			if err != nil {
				return
			}
		case "oracle":
			err = dc.ReadExtension(&z.Oracle)
			if err != nil {
				return
			}
		case "total":
			err = dc.ReadExtension(&z.Total)
			if err != nil {
				return
			}
		case "hash":
			err = dc.ReadExtension(&z.Hash)
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
func (z *Benefit) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "balance"
	err = en.Append(0x87, 0xa7, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Balance)
	if err != nil {
		return
	}
	// write "vote"
	err = en.Append(0xa4, 0x76, 0x6f, 0x74, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Vote)
	if err != nil {
		return
	}
	// write "network"
	err = en.Append(0xa7, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Network)
	if err != nil {
		return
	}
	// write "storage"
	err = en.Append(0xa7, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Storage)
	if err != nil {
		return
	}
	// write "oracle"
	err = en.Append(0xa6, 0x6f, 0x72, 0x61, 0x63, 0x6c, 0x65)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Oracle)
	if err != nil {
		return
	}
	// write "total"
	err = en.Append(0xa5, 0x74, 0x6f, 0x74, 0x61, 0x6c)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Total)
	if err != nil {
		return
	}
	// write "hash"
	err = en.Append(0xa4, 0x68, 0x61, 0x73, 0x68)
	if err != nil {
		return
	}
	err = en.WriteExtension(&z.Hash)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Benefit) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "balance"
	o = append(o, 0x87, 0xa7, 0x62, 0x61, 0x6c, 0x61, 0x6e, 0x63, 0x65)
	o, err = msgp.AppendExtension(o, &z.Balance)
	if err != nil {
		return
	}
	// string "vote"
	o = append(o, 0xa4, 0x76, 0x6f, 0x74, 0x65)
	o, err = msgp.AppendExtension(o, &z.Vote)
	if err != nil {
		return
	}
	// string "network"
	o = append(o, 0xa7, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b)
	o, err = msgp.AppendExtension(o, &z.Network)
	if err != nil {
		return
	}
	// string "storage"
	o = append(o, 0xa7, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65)
	o, err = msgp.AppendExtension(o, &z.Storage)
	if err != nil {
		return
	}
	// string "oracle"
	o = append(o, 0xa6, 0x6f, 0x72, 0x61, 0x63, 0x6c, 0x65)
	o, err = msgp.AppendExtension(o, &z.Oracle)
	if err != nil {
		return
	}
	// string "total"
	o = append(o, 0xa5, 0x74, 0x6f, 0x74, 0x61, 0x6c)
	o, err = msgp.AppendExtension(o, &z.Total)
	if err != nil {
		return
	}
	// string "hash"
	o = append(o, 0xa4, 0x68, 0x61, 0x73, 0x68)
	o, err = msgp.AppendExtension(o, &z.Hash)
	if err != nil {
		return
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Benefit) UnmarshalMsg(bts []byte) (o []byte, err error) {
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
		case "balance":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Balance)
			if err != nil {
				return
			}
		case "vote":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Vote)
			if err != nil {
				return
			}
		case "network":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Network)
			if err != nil {
				return
			}
		case "storage":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Storage)
			if err != nil {
				return
			}
		case "oracle":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Oracle)
			if err != nil {
				return
			}
		case "total":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Total)
			if err != nil {
				return
			}
		case "hash":
			bts, err = msgp.ReadExtensionBytes(bts, &z.Hash)
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
func (z *Benefit) Msgsize() (s int) {
	s = 1 + 8 + msgp.ExtensionPrefixSize + z.Balance.Len() + 5 + msgp.ExtensionPrefixSize + z.Vote.Len() + 8 + msgp.ExtensionPrefixSize + z.Network.Len() + 8 + msgp.ExtensionPrefixSize + z.Storage.Len() + 7 + msgp.ExtensionPrefixSize + z.Oracle.Len() + 6 + msgp.ExtensionPrefixSize + z.Total.Len() + 5 + msgp.ExtensionPrefixSize + z.Hash.Len()
	return
}
